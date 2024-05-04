package reader

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type OCIRegistry struct {
	logger   *log.Logger
	username string
	password string
}

func NewOCIRegistry(logger *log.Logger, username, password string) *OCIRegistry {
	return &OCIRegistry{
		logger:   logger,
		username: username,
		password: password,
	}
}

// PullFromRegistry loads an image from a remote OCI registry
func (r *OCIRegistry) Pull(imageRef string) (v1.Image, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		panic(err)
	}

	b := authn.Basic{
		Username: r.username,
		Password: r.password,
	}

	cfg, err := b.Authorization()
	if err != nil {
		return nil, err
	}

	auth := authn.FromConfig(*cfg)

	return remote.Image(ref, remote.WithAuth(auth), remote.WithProgress(r.progressReport()))
}

func (r *OCIRegistry) progressReport() chan v1.Update {
	ch := make(chan v1.Update, 1)
	total := int64(0)
	completed := int64(0)

	t := time.AfterFunc(5*time.Second, func() {
		percentage := "0%"

		if completed > 0 && total > 0 {
			p := int((completed / total) * 100)

			percentage = fmt.Sprintf("%d%%", p)

			r.logger.Info("Pulling image", "complete", percentage, "total", total, "completed", completed)
		}
	})

	go func() {
		for {
			update, ok := <-ch
			if !ok {
				r.logger.Info("Pull image complete")
				t.Stop()
				return
			}

			total = update.Total
			completed = update.Complete
		}
	}()

	return ch
}
