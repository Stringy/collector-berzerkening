package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
)

type CollectorManager struct {
	executor        Executor
	Mounts          map[string]string
	Env             map[string]string
	CollectorOutput string
	CollectorImage  string
	BootstrapOnly   bool
	TestName        string
	CoreDumpFile    string
	VmConfig        string
	ContainerID     string
}

func NewCollectorManager(e Executor, name string) *CollectorManager {
	collectorConfig := `{"logLevel":"info","turnOffScrape":true,"scrapeInterval":2}`

	env := map[string]string{
		"GRPC_SERVER":                     "192.168.56.1:1338",
		"COLLECTOR_CONFIG":                collectorConfig,
		"COLLECTION_METHOD":               "core-bpf",
		"COLLECTOR_PRE_ARGUMENTS":         "",
		"ENABLE_CORE_DUMP":                "false",
		"ROX_COLLECTOR_CORE_BPF_HARDFAIL": "true",
		"MODULE_DOWNLOAD_BASE_URL":        "https://collector-modules.stackrox.io/612dd2ee06b660e728292de9393e18c81a88f347ec52a39207c5166b5302b656",
	}

	mounts := map[string]string{
		// The presence of this socket disables an optimisation, which would turn off podman runtime parsing.
		// https://github.com/falcosecurity/libs/pull/296
		"/run/podman/podman.sock:ro": "/var/run/docker.sock",
		"/host/proc:ro":              "/proc",
		"/host/etc:ro":               "/etc/",
		"/host/usr/lib:ro":           "/usr/lib/",
		"/host/sys:ro":               "/sys/",
		"/host/dev:ro":               "/dev",
		"/tmp":                       "/tmp",
		// /module is an anonymous volume to reflect the way collector
		// is usually run in kubernetes (with in-memory volume for /module)
		"/module": "",
	}

	return &CollectorManager{
		executor:       e,
		BootstrapOnly:  false,
		CollectorImage: "quay.io/stackrox-io/collector:3.16.0",
		Env:            env,
		Mounts:         mounts,
		TestName:       name,
		CoreDumpFile:   "/tmp/core.out",
		VmConfig:       "game",
	}
}

func (c *CollectorManager) Setup() error {
	return c.executor.PullImage(c.CollectorImage)
}

func (c *CollectorManager) Launch() error {
	return c.launchCollector()
}

func (c *CollectorManager) TearDown() error {
	isRunning, err := c.executor.IsContainerRunning("collector")
	if err != nil {
		fmt.Println("Error: Checking if container running")
		return err
	}

	if !isRunning {
		c.captureLogs("collector")
		// Check if collector container segfaulted or exited with error
		exitCode, err := c.executor.ExitCode("collector")
		if err != nil {
			fmt.Println("Error: Container not running")
			return err
		}
		if exitCode != 0 {
			return fmt.Errorf("Collector container has non-zero exit code (%d)", exitCode)
		}
	} else {
		c.stopContainer("collector")
		c.captureLogs("collector")
		c.killContainer("collector")
	}

	return nil
}

// These two methods might be useful in the future. I used them for debugging
func (c *CollectorManager) getContainers() (string, error) {
	cmd := []string{RuntimeCommand, "container", "ps"}
	containers, err := c.executor.Exec(cmd...)

	return containers, err
}

func (c *CollectorManager) getAllContainers() (string, error) {
	cmd := []string{RuntimeCommand, "container", "ps", "-a"}
	containers, err := c.executor.Exec(cmd...)

	return containers, err
}

func (c *CollectorManager) launchCollector() error {
	cmd := []string{RuntimeCommand, "run",
		"--name", "collector",
		"--privileged",
		"--network=host"}

	if !c.BootstrapOnly {
		cmd = append(cmd, "-d")
	}

	for dst, src := range c.Mounts {
		mount := src + ":" + dst
		if src == "" {
			// allows specification of anonymous volumes
			mount = dst
		}
		cmd = append(cmd, "-v", mount)
	}
	for k, v := range c.Env {
		cmd = append(cmd, "--env", k+"="+v)
	}

	cmd = append(cmd, c.CollectorImage)

	if c.BootstrapOnly {
		cmd = append(cmd, "exit", "0")
	}

	output, err := c.executor.Exec(cmd...)
	c.CollectorOutput = output

	outLines := strings.Split(output, "\n")
	c.ContainerID = ContainerShortID(string(outLines[len(outLines)-1]))
	return err
}

func (c *CollectorManager) captureLogs(containerName string) (string, error) {
	logs, err := c.executor.Exec(RuntimeCommand, "logs", containerName)
	if err != nil {
		fmt.Printf(RuntimeCommand+" logs error (%v) for container %s\n", err, containerName)
		return "", err
	}
	logDirectory := filepath.Join(".", "container-logs", c.VmConfig, c.Env["COLLECTION_METHOD"])
	os.MkdirAll(logDirectory, os.ModePerm)
	logFile := filepath.Join(logDirectory, strings.ReplaceAll(c.TestName, "/", "_")+"-"+containerName+".log")
	err = ioutil.WriteFile(logFile, []byte(logs), 0644)
	if err != nil {
		return "", err
	}
	return logs, nil
}

func (c *CollectorManager) killContainer(name string) error {
	_, err1 := c.executor.Exec(RuntimeCommand, "kill", name)
	_, err2 := c.executor.Exec(RuntimeCommand, "rm", "-fv", name)

	var result error
	if err1 != nil {
		result = multierror.Append(result, err1)
	}
	if err2 != nil {
		result = multierror.Append(result, err2)
	}

	return result
}

func (c *CollectorManager) stopContainer(name string) error {
	_, err := c.executor.Exec(RuntimeCommand, "stop", name)
	return err
}
