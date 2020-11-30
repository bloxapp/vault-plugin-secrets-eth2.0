package launcher

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
)

var buildOnce sync.Once

// Config contains configuration of validator service instance.
type Config struct {
	ID    string
	URL   string
	Token string
}

// Docker implements the logic to launch test container.
type Docker struct {
	client    *client.Client
	imageName string
	basePath  string
}

// New is the constructor of dockerLauncher
func New(imageName, basePath string) (*Docker, error) {
	// Create a new client
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create docker client")
	}

	return &Docker{
		client:    cli,
		imageName: imageName,
		basePath:  basePath,
	}, nil
}

// Launch implements launcher.Launcher interface by starting image using installed Docker service.
func (l *Docker) Launch(ctx context.Context, name string) (*Config, error) {
	var buildErr error
	buildOnce.Do(func() {
		buildErr = l.buildImage(ctx)
	})
	if buildErr != nil {
		return nil, errors.Wrap(buildErr, "failed to build image")
	}

	// Get available port
	hostPort, err := getFreePort()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get available port")
	}

	containerPort, err := nat.NewPort("tcp", "8200")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get the port")
	}

	portBinding := nat.PortMap{
		containerPort: []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: strconv.Itoa(hostPort),
		}},
	}
	cont, err := l.client.ContainerCreate(
		ctx,
		&container.Config{
			Image: l.imageName,
			Env: []string{
				"VAULT_EXTERNAL_ADDRESS=172.17.0.1",
				"VAULT_CLIENT_TIMEOUT=30s",
				"UNSEAL=true",
			},
		},
		&container.HostConfig{
			PortBindings: portBinding,
			CapAdd:       strslice.StrSlice{"IPC_LOCK"},
		}, nil, name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	// Start container
	if err := l.client.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{}); err != nil {
		return nil, errors.Wrap(err, "failed to start container")
	}

	// Retrieve container config
	containerConfig, err := l.client.ContainerInspect(ctx, cont.ID)
	if err != nil {
		l.Stop(ctx, cont.ID)
		return nil, errors.Wrapf(err, "failed to inspect container with ID %s", cont.ID)
	}

	// Retrieve container IP address
	var ip string
	for _, network := range containerConfig.NetworkSettings.Networks {
		if len(network.Gateway) > 0 {
			ip = network.Gateway
			break
		}
	}
	ip = "127.0.0.1"

	// Read logs so we can understand the plugin is loaded
	logsStream, err := l.client.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		l.Stop(ctx, cont.ID)
		return nil, errors.Wrap(err, "failed to read logs")
	}
	defer logsStream.Close()

	// Read logs from stream
	hdr := make([]byte, 8)
	timeout := time.Tick(time.Minute * 3)
	for {
		var loaded bool
		select {
		case <-timeout:
			return nil, errors.New("failed to read logs: time out")
		default:
			if _, err := logsStream.Read(hdr); err != nil {
				l.Stop(ctx, cont.ID)
				return nil, errors.Wrap(err, "failed to read from logs stream")
			}

			count := binary.BigEndian.Uint32(hdr[4:])
			dat := make([]byte, count)
			_, err = logsStream.Read(dat)
			if err != nil {
				l.Stop(ctx, cont.ID)
				return nil, errors.Wrap(err, "failed to read from logs stream")
			}

			dta := strings.ToLower(string(dat))
			fmt.Println("dta", dta)
			if strings.Contains(dta, "connection refused") {
				l.Stop(ctx, cont.ID)
				return nil, errors.Errorf("failed to launch instance: %s", string(dat))
			} else if strings.Contains(dta, "core: successfully reloaded plugin: plugin=ethsign path=ethereum") {
				loaded = true
				break
			}
		}

		if loaded {
			break
		}
	}

	// Retrieve auth root token
	tokenData, err := l.inspectExecResp(ctx, cont.ID, []string{"cat", "/data/keys/vault.root.token"})
	if err != nil {
		l.Stop(ctx, cont.ID)
		return nil, errors.Wrap(err, "failed to read exec command of reading token result")
	} else if tokenData.exitCode != 0 || len(tokenData.stdErr) > 0 {
		l.Stop(ctx, cont.ID)
		return nil, errors.Errorf("failed to retrieve auth token: exit code %d, error %s", tokenData.exitCode, tokenData.stdErr)
	}

	return &Config{
		ID:    cont.ID,
		URL:   fmt.Sprintf("https://%s:%d", ip, hostPort),
		Token: strings.TrimSpace(tokenData.stdOut),
	}, nil
}

// Stop stops the docker container by the given ID.
func (l *Docker) Stop(ctx context.Context, id string) error {
	// Stop container
	if err := l.client.ContainerStop(ctx, id, nil); err != nil {
		return errors.Wrap(err, "failed to stop container")
	}

	return nil
}

// buildImage builds the test image
func (l *Docker) buildImage(ctx context.Context) error {
	buildCtx, err := archive.TarWithOptions(l.basePath, &archive.TarOptions{})
	if err != nil {
		return err
	}

	imageBuildResponse, err := l.client.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Context:    buildCtx,
		Dockerfile: "Dockerfile",
		Remove:     true,
		Tags:       []string{l.imageName},
	})
	if err != nil {
		return err
	}
	defer imageBuildResponse.Body.Close()

	_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
	if err != nil {
		return err
	}

	return nil
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
