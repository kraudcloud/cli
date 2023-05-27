package compose

import (
	"context"
	"encoding/json"
	"io"

	"github.com/nofeaturesonlybugs/set"

	"github.com/compose-spec/compose-go/types"
	dtype "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	dc "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type UpOptions struct {
	level logrus.Level
	force bool
}

func Up(ctx context.Context, dcc *dc.Client, project *types.Project, options ...func(*UpOptions)) error {
	log := logrus.New().WithFields(logrus.Fields{
		"project": coalesce(project.Name, project.WorkingDir),
	}).WithContext(ctx)

	o := &UpOptions{}
	for _, option := range options {
		option(o)
	}

	log.Logger.SetLevel(o.level)

	existVols, err := dcc.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return err
	}

	// TODO: make it right (aka use the right networks and stuff, currently we just ignore them)

conVol:
	for volName, createVol := range project.Volumes {
		log.Trace(js(createVol))

		// check if volume exists
		for i := range existVols.Volumes {
			eVol := existVols.Volumes[i]
			if volName == eVol.Name {
				log.Debugf("volume %s already exists", volName)
				continue conVol
			}
		}

		log.Debugf("create volume: %s", volName)
		_, err := dcc.VolumeCreate(ctx, volume.CreateOptions{
			Driver:     createVol.Driver,
			DriverOpts: createVol.DriverOpts,
			Labels:     createVol.Labels,
			Name:       volName,
		})
		if err != nil {
			return err
		}
	}

	err = project.WithServices(project.ServiceNames(), func(svc types.ServiceConfig) error {
		log.Trace(js(svc))

		r, _, err := dcc.ImageInspectWithRaw(ctx, svc.Image)
		if r.ID == "" || err != nil {
			log.Debugf("pull image: %s", svc.Image)
			r, err := dcc.ImagePull(ctx, svc.Image, dtype.ImagePullOptions{
				RegistryAuth: "null",
			})
			if err != nil {
				return err
			}
			defer r.Close()
			io.Copy(io.Discard, r)
		}

		// check if container exists
		_, err = dcc.ContainerInspect(ctx, svc.Name)
		if err == nil {
			log.Debugf("container %s already exists", svc.Name)

			if o.force {
				log.Infof("force-removing container: %s", svc.Name)
				err = dcc.ContainerRemove(ctx, svc.Name, dtype.ContainerRemoveOptions{
					Force: true,
				})
				if err != nil {
					return err
				}
			}
		}

		log.Debugf("create container: %s", svc.Name)
		c, err := dcc.ContainerCreate(
			ctx,
			castBestEffort[*container.Config](svc),
			nil,
			nil,
			nil,
			svc.Name,
		)
		if err != nil {
			return err
		}

		log.Debugf("start container: %s", svc.Name)
		err = dcc.ContainerStart(ctx, c.ID, dtype.ContainerStartOptions{})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func WithLevel(l logrus.Level) func(*UpOptions) {
	return func(o *UpOptions) {
		o.level = l
	}
}

func WithForce() func(*UpOptions) {
	return func(o *UpOptions) {
		o.force = true
	}
}

func coalesce[T comparable](v ...T) T {
	var zero T
	for _, i := range v {
		if i != zero {
			return i
		}
	}

	return zero
}

func js(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(b)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func castBestEffort[T any](v any) T {
	out := new(T)
	set.V(out).Fill(reflectMapper(v))
	return *out
}

func reflectMapper[T any](v T) set.Getter {
	vv := set.V(v)
	if vv.IsMap {
		return set.MapGetter(v)
	}

	return set.GetterFunc(func(name string) any {
		v := vv.TopValue.FieldByName(name)
		if !v.IsValid() {
			return nil
		}

		return v.Interface()
	})
}
