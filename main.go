package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"
)

var (
	log             *logrus.Logger
	verbose         bool
	cgroupHostPath  string
	podLabelSelctor string
)

func matchLabelSelector(s string, podLabels map[string]string) bool {
	if podLabels == nil {
		return false
	}
	selector, err := labels.Parse(s)
	if err != nil {
		return false
	}
	return selector.Matches(labels.Set(podLabels))
}

// Dump one or more objects, with an optional global prefix and per-object tags.
func dump(args ...interface{}) {
	var (
		prefix string
		idx    int
	)

	if len(args)&0x1 == 1 {
		prefix = args[0].(string)
		idx++
	}

	for ; idx < len(args)-1; idx += 2 {
		tag, obj := args[idx], args[idx+1]
		msg, err := yaml.Marshal(obj)
		if err != nil {
			log.Infof("%s: %s: failed to dump object: %v", prefix, tag, err)
			continue
		}

		if prefix != "" {
			log.Infof("%s: %s:", prefix, tag)
			for _, line := range strings.Split(strings.TrimSpace(string(msg)), "\n") {
				log.Infof("%s:    %s", prefix, line)
			}
		} else {
			log.Infof("%s:", tag)
			for _, line := range strings.Split(strings.TrimSpace(string(msg)), "\n") {
				log.Infof("  %s", line)
			}
		}
	}
}

// our injector plugin
type plugin struct {
	stub stub.Stub
}

func (p *plugin) modifyPodOOMGroup(_ context.Context, pod *api.PodSandbox) error {
	if verbose {
		dump("modifyPodOOMGroup", "pod", pod)
	}
	if pod.Linux == nil {
		log.Warnf("pod %q has no Linux configuration, skipping OOM group modification", pod.Id)
		return nil
	}
	err := filepath.WalkDir(cgroupHostPath+pod.Linux.CgroupParent, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "memory.oom.group" {
			err := os.WriteFile(path, []byte("0"), 0644)
			return err
		}
		return nil
	})
	return err
}
func (p *plugin) StartContainer(ctx context.Context, pod *api.PodSandbox, ctr *api.Container) error {
	if pod == nil || ctr == nil {
		return nil
	}
	if podLabelSelctor == "" {
		if err := p.modifyPodOOMGroup(ctx, pod); err != nil {
			log.Errorf("failed to modify pod %q OOM group: %v", pod.Id, err)
			return err
		}
		return nil
	}
	if pod.Labels != nil || !matchLabelSelector(podLabelSelctor, pod.Labels) {
		if verbose {
			log.Infof("pod %q does not match label selector %q, skipping", pod.Id, podLabelSelctor)
		}
	}
	return nil
}

func (p *plugin) PostUpdateContainer(ctx context.Context, pod *api.PodSandbox, ctr *api.Container) error {
	if pod == nil || ctr == nil {
		return nil
	}
	if podLabelSelctor == "" {
		if err := p.modifyPodOOMGroup(ctx, pod); err != nil {
			log.Errorf("failed to modify pod %q OOM group: %v", pod.Id, err)
			return err
		}
		return nil
	}
	if !matchLabelSelector(podLabelSelctor, pod.Labels) {
		if verbose {
			log.Infof("pod %q does not match label selector %q, skipping", pod.Id, podLabelSelctor)
		}
		return nil
	}
	if err := p.modifyPodOOMGroup(ctx, pod); err != nil {
		log.Errorf("failed to modify pod %q OOM group: %v", pod.Id, err)
		return err
	}
	return nil
}

func main() {
	var (
		pluginName string
		pluginIdx  string
		opts       []stub.Option
		err        error
	)
	log = logrus.StandardLogger()
	log.SetFormatter(&logrus.TextFormatter{
		PadLevelText: true,
	})

	flag.StringVar(&pluginName, "name", "", "plugin name to register to NRI")
	flag.StringVar(&pluginIdx, "idx", "", "plugin index to register to NRI")
	flag.BoolVar(&verbose, "verbose", false, "enable (more) verbose logging")
	flag.StringVar(&cgroupHostPath, "cgroup-path", "/host-sys/fs/cgroup", "source host path for cgroup mounts")
	flag.StringVar(&podLabelSelctor, "label-selector", "app=zestu", "label selector to filter pods for which the plugin should be invoked")
	flag.Parse()
	if podLabelSelctor != "" {
		if _, err := labels.Parse(podLabelSelctor); err != nil {
			log.Fatalf("failed to parse label selector %q: %v", podLabelSelctor, err)
		}
	}
	if pluginName != "" {
		opts = append(opts, stub.WithPluginName(pluginName))
	}
	if pluginIdx != "" {
		opts = append(opts, stub.WithPluginIdx(pluginIdx))
	}

	p := &plugin{}
	if p.stub, err = stub.New(p, opts...); err != nil {
		log.Fatalf("failed to create plugin stub: %v", err)
	}

	err = p.stub.Run(context.Background())
	if err != nil {
		log.Errorf("plugin exited with error %v", err)
		os.Exit(1)
	}
}
