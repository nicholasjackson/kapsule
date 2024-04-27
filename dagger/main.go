package main

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/charmbracelet/log"
)

var oses = []string{"linux", "darwin", "windows"}
var arches = []string{"amd64", "arm64"}
var owner = "kapsule"
var repo = "nicholasjackson"

func New() *Kapsule {
	return &Kapsule{}
}

type Kapsule struct {
	lastError     error
	goCacheVolume *CacheVolume
}

func (d *Kapsule) WithGoCache(cache *CacheVolume) *Kapsule {
	d.goCacheVolume = cache
	return d
}

func (d *Kapsule) All(
	ctx context.Context,
	src *Directory,
	// +optional
	quick bool,
	// +optional
	githubToken *Secret,
	// +optional
	notorizeCert *File,
	// +optional
	notorizeCertPassword *Secret,
	// +optional
	notorizeKey *File,
	// +optional
	notorizeId string,
	// +optional
	notorizeIssuer string,
) (*Directory, error) {
	// if quick, only build for the current architecture
	if quick {
		d.setArchLocalMachine(ctx)
	}

	// get the version
	version := "0.0.0"
	sha := ""

	var output *Directory

	// remove the build output directory from the source
	src = src.
		WithoutDirectory(".dagger").
		WithoutDirectory("build-output")

	// if we have a github token, get the version from the associated PR label
	if githubToken != nil {
		version, sha, _ = d.getVersion(ctx, githubToken, src)
	}

	log.Info("Building version", "semver", version, "sha", sha)

	// run the unit tests
	d.UnitTest(ctx, src, !quick)

	// build the applications
	output, _ = d.Build(ctx, src, version, sha)

	// package the build outputs
	output, _ = d.Package(ctx, output, version)

	// create the archives
	output, _ = d.Archive(ctx, output, version)

	// if we have the notorization details sign and notorize the osx binaries
	if notorizeCert != nil && notorizeCertPassword != nil && notorizeKey != nil && notorizeId != "" && notorizeIssuer != "" {
		output, _ = d.SignAndNotorize(ctx, version, output, notorizeCert, notorizeCertPassword, notorizeKey, notorizeId, notorizeIssuer)
	}

	// generate the checksums
	output, _ = d.GenerateChecksums(ctx, output, version)

	return output, d.lastError
}

func (d *Kapsule) Release(
	ctx context.Context,
	src *Directory,
	archives *Directory,
	githubToken *Secret,
	gemfuryToken *Secret,
) (string, error) {
	// create a new github release
	version, _ := d.GithubRelease(ctx, src, archives, githubToken)

	// update the brew formula at kapsule-labs/homebrew-repo
	d.UpdateBrew(ctx, version, githubToken)

	//	update the gemfury repository
	d.UpdateGemFury(ctx, version, gemfuryToken, archives)

	return version, d.lastError
}

func (d *Kapsule) Build(
	ctx context.Context,
	src *Directory,
	version,
	sha string,
) (*Directory, error) {
	if d.hasError() {
		return nil, d.lastError
	}

	cli := dag.Pipeline("build")

	// create empty directory to put build outputs
	outputs := cli.Directory()

	// get `golang` image
	golang := cli.Container().
		From("golang:latest").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithMountedCache("/go/pkg/mod", d.goCache())

	for _, goos := range oses {
		for _, goarch := range arches {
			fmt.Println("Build for", goos, goarch, "...")

			// create a directory for each os and arch
			path := fmt.Sprintf("build/%s/%s/", goos, goarch)

			// set GOARCH and GOOS in the build environment
			build, err := golang.
				WithEnvVariable("CGO_ENABLED", "0").
				WithEnvVariable("GOOS", goos).
				WithEnvVariable("GOARCH", goarch).
				WithExec([]string{
					"go", "build",
					"-o", path,
					"-ldflags", fmt.Sprintf("-X main.version=%s -X main.sha=%s", version, sha),
					"./cmd",
				}).
				Sync(ctx)

			if err != nil {
				d.lastError = err
				return nil, err
			}

			// get reference to build output directory in container
			outputs = outputs.WithDirectory(path, build.Directory(path))
		}
	}

	return outputs, nil
}

func (d *Kapsule) UnitTest(
	ctx context.Context,
	src *Directory,
	withRace bool,
) error {
	if d.hasError() {
		return d.lastError
	}

	cli := dag.Pipeline("unit-test")

	raceFlag := ""
	if withRace {
		raceFlag = "-race"
	}

	golang := cli.Container().
		From("golang:latest").
		WithDirectory("/src", src).
		WithMountedCache("/go/pkg/mod", d.goCache()).
		WithWorkdir("/src").
		WithExec([]string{"go", "test", "-v", raceFlag, "./..."})

	_, err := golang.Sync(ctx)
	if err != nil {
		d.lastError = err
	}

	return err
}

func (d *Kapsule) Package(
	ctx context.Context,
	binaries *Directory,
	version string,
) (*Directory, error) {
	if d.hasError() {
		return nil, d.lastError
	}

	cli := dag.Pipeline("package")

	for _, os := range oses {
		if os == "linux" {
			for _, a := range arches {
				// create a package directory including the binaries
				pkg := cli.Directory()
				pkg = pkg.WithFile("/bin/kapsule", binaries.File(fmt.Sprintf("/build/linux/%s/kapsule", a)))

				// create a debian package
				p := cli.Deb().Build(pkg, a, "kapsule", version, "Nic Jackson", "Kapsule application")

				// add the debian package to the binaries directory
				binaries = binaries.WithFile(fmt.Sprintf("/pkg/linux/%s/kapsule.deb", a), p)
			}
		}
	}

	return binaries, nil
}

type Archive struct {
	// path of the
	Path   string
	Type   string
	Output string
}

var archives = []Archive{
	{Path: "/build/windows/amd64/kapsule.exe", Type: "zip", Output: "kapsule_%%VERSION%%_windows_x86_64.zip"},
	{Path: "/build/darwin/amd64/kapsule", Type: "zip", Output: "kapsule_%%VERSION%%_darwin_x86_64.zip"},
	{Path: "/build/darwin/arm64/kapsule", Type: "zip", Output: "kapsule_%%VERSION%%_darwin_arm64.zip"},
	{Path: "/build/linux/amd64/kapsule", Type: "targz", Output: "kapsule_%%VERSION%%_linux_x86_64.tar.gz"},
	{Path: "/build/linux/arm64/kapsule", Type: "targz", Output: "kapsule_%%VERSION%%_linux_arm64.tar.gz"},
	{Path: "/pkg/linux/amd64/kapsule.deb", Type: "copy", Output: "kapsule_%%VERSION%%_linux_x86_64.deb"},
	{Path: "/pkg/linux/arm64/kapsule.deb", Type: "copy", Output: "kapsule_%%VERSION%%_linux_arm64.deb"},
}

// Archive creates zipped and tar archives of the binaries
func (d *Kapsule) Archive(
	ctx context.Context,
	binaries *Directory,
	version string,
) (*Directory, error) {
	if d.hasError() {
		return nil, d.lastError
	}

	cli := dag.Pipeline("archive")
	out := cli.Directory()

	zipContainer := cli.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "zip"})

	checksums := strings.Builder{}

	for _, a := range archives {
		outPath := strings.ReplaceAll(a.Output, "%%VERSION%%", version)
		switch a.Type {
		case "zip":
			// create a zip archive

			// first get the filename as with windows this has an extension
			fn := path.Base(a.Path)

			// zip the file
			zip := zipContainer.
				WithMountedFile(fn, binaries.File(a.Path)).
				WithExec([]string{"zip", "-r", outPath, fn})

			out = out.WithFile(outPath, zip.File(outPath))
		case "targz":
			// create a zip archive
			zip := zipContainer.
				WithMountedFile("/kapsule", binaries.File(a.Path)).
				WithExec([]string{"tar", "-czf", outPath, "/kapsule"})

			out = out.WithFile(outPath, zip.File(outPath))
		case "copy":
			out = out.WithFile(outPath, binaries.File(a.Path))
		}

		// generate the checksum
		cs, err := cli.Checksum().CalculateFromFile(ctx, out.File(outPath))
		if err != nil {
			d.lastError = fmt.Errorf("unable to generate checksum for archive: %w", err)
			return nil, d.lastError
		}

		// checksum is returned as "checksum filename" we need to remove the filename as it is not
		// the same as the release name
		csParts := strings.Split(cs, " ")

		checksums.WriteString(fmt.Sprintf("%s  %s\n", csParts[0], outPath))
	}

	out = out.WithNewFile("checksums.txt", checksums.String())

	return out, nil
}

func (d Kapsule) GenerateChecksums(
	ctx context.Context,
	files *Directory,
	version string,
) (*Directory, error) {
	cli := dag.Pipeline("generate-checksums")
	checksums := strings.Builder{}

	for _, a := range archives {
		outPath := strings.ReplaceAll(a.Output, "%%VERSION%%", version)

		// generate the checksum
		cs, err := cli.Checksum().CalculateFromFile(ctx, files.File(outPath))
		if err != nil {
			d.lastError = fmt.Errorf("unable to generate checksum for archive: %w", err)
			return nil, d.lastError
		}

		// checksum is returned as "checksum filename" we need to remove the filename as it is not
		// the same as the release name
		csParts := strings.Split(cs, " ")

		checksums.WriteString(fmt.Sprintf("%s  %s\n", csParts[0], outPath))
	}

	files = files.WithNewFile("checksums.txt", checksums.String())
	return files, nil
}

var notorize = []Archive{
	{Path: "/kapsule_%%VERSION%%_darwin_x86_64.zip", Type: "zip", Output: "/kapsule_%%VERSION%%_darwin_x86_64.zip"},
	{Path: "/kapsule_%%VERSION%%_darwin_arm64.zip", Type: "zip", Output: "/kapsule_%%VERSION%%_darwin_arm64.zip"},
}

// SignAndNotorize signs and notorizes the osx binaries using the Apple notary service
func (d Kapsule) SignAndNotorize(
	ctx context.Context,
	version string,
	archives *Directory,
	cert *File,
	password *Secret,
	key *File,
	keyId,
	keyIssuer string,
) (*Directory, error) {
	if d.hasError() {
		return nil, d.lastError
	}

	cli := dag.Pipeline("notorize")

	not := dag.Notorize().
		WithP12Cert(cert, password).
		WithNotoryKey(key, keyId, keyIssuer)

	out := archives

	zipContainer := cli.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "zip"})

	for _, a := range notorize {
		path := strings.ReplaceAll(a.Output, "%%VERSION%%", version)

		jpFile := zipContainer.
			WithMountedFile("/kapsule.zip", archives.File(path)).
			WithExec([]string{"unzip", "/kapsule.zip"}).
			File("/kapsule")

		notorized := not.SignAndNotorize(jpFile)

		nFile := zipContainer.
			WithMountedFile("/kapsule", notorized).
			WithExec([]string{"zip", "-r", path, "/kapsule"}).
			File(path)

		out = out.WithFile(path, nFile)
	}

	return out, nil
}

func (d *Kapsule) GithubRelease(
	ctx context.Context,
	src *Directory,
	archives *Directory,
	githubToken *Secret,
) (string, error) {
	if d.hasError() {
		return "", d.lastError
	}

	version, sha, err := d.getVersion(ctx, githubToken, src)
	if err != nil {
		d.lastError = err
		return "", err
	}

	if version == "0.0.0" {
		d.lastError = fmt.Errorf("no version to release, did you tag the PR?")
		return "", d.lastError
	}

	cli := dag.Pipeline("release")

	_, err = cli.Github().
		WithToken(githubToken).
		CreateRelease(ctx, owner, repo, version, sha, GithubCreateReleaseOpts{Files: archives})

	if err != nil {
		d.lastError = err
		return "", err
	}

	return version, err
}

func (d *Kapsule) UpdateBrew(
	ctx context.Context,
	version string,
	githubToken *Secret,
) error {
	if d.hasError() {
		return d.lastError
	}

	cli := dag.Pipeline("update-brew")

	_, err := cli.Brew().Formula(
		ctx,
		"https://github.com/nicholasjackson/kapsule",
		"nicholasjackson/homebrew-repo",
		version,
		"Nic Jackson",
		"jackson.nic@gmail.com",
		"kapsule",
		githubToken,
		BrewFormulaOpts{
			DarwinX86Url:   fmt.Sprintf("https://github.com/kapsule-labs/kapsule/releases/download/%s/kapsule_%s_darwin_x86_64.zip", version, version),
			DarwinArm64Url: fmt.Sprintf("https://github.com/kapsule-labs/kapsule/releases/download/%s/kapsule_%s_darwin_arm64.zip", version, version),
			LinuxX86Url:    fmt.Sprintf("https://github.com/kapsule-labs/kapsule/releases/download/%s/kapsule_%s_linux_x86_64.tar.g", version, version),
			LinuxArm64Url:  fmt.Sprintf("https://github.com/kapsule-labs/kapsule/releases/download/%s/kapsule_%s_linux_arm64.tar.giz", version, version),
		},
	)

	if err != nil {
		d.lastError = err
	}

	return err
}

var gemFury = []Archive{
	{Path: "/pkg/linux/amd64/kapsule.deb", Type: "copy", Output: "kapsule_%%VERSION%%_linux_x86_64.deb"},
	{Path: "/pkg/linux/arm64/kapsule.deb", Type: "copy", Output: "kapsule_%%VERSION%%_linux_arm64.deb"},
}

func (d *Kapsule) UpdateGemFury(
	ctx context.Context,
	version string,
	gemFuryToken *Secret,
	archives *Directory,
) error {
	cli := dag.Pipeline("update-gem-fury")

	tkn, _ := gemFuryToken.Plaintext(ctx)
	url := fmt.Sprintf("https://%s@push.fury.io/nicholasjackson/", tkn)

	for _, a := range gemFury {
		output := strings.Replace(a.Output, "%%VERSION%%", version, 1)

		_, err := cli.Container().
			From("curlimages/curl:latest").
			WithFile(output, archives.File(output)).
			WithExec([]string{"-F", fmt.Sprintf("package=@%s", output), url}).
			Sync(ctx)

		if err != nil {
			d.lastError = err
			return err
		}
	}

	return nil
}

func (d *Kapsule) getVersion(ctx context.Context, token *Secret, src *Directory) (string, string, error) {
	if d.hasError() {
		return "", "", d.lastError
	}

	cli := dag.Pipeline("get-version")

	// get the latest git sha from the source
	ref, err := cli.Container().
		From("alpine/git").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"rev-parse", "HEAD"}).
		Stdout(ctx)

	if err != nil {
		d.lastError = err
		return "", "", err
	}

	// make sure there is no whitespace from the output
	ref = strings.TrimSpace(ref)
	log.Info("github reference", "sha", ref)

	// get the next version from the associated PR label
	v, err := cli.Github().
		WithToken(token).
		NextVersionFromAssociatedPrlabel(ctx, owner, repo, ref)

	if err != nil {
		d.lastError = err
		return "", "", err
	}

	// if there is no version, default to 0.0.0
	if v == "" {
		v = "0.0.0"
	}

	log.Info("new version", "semver", v)

	return v, ref, nil
}

func (d *Kapsule) goCache() *CacheVolume {
	if d.goCacheVolume == nil {
		d.goCacheVolume = dag.CacheVolume("go-cache")
	}

	return d.goCacheVolume
}

func (d *Kapsule) hasError() bool {
	return d.lastError != nil
}

func (d *Kapsule) setArchLocalMachine(ctx context.Context) {
	// get the architecture of the current machine
	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		panic(err)
	}

	arch := strings.Split(string(platform), "/")[1]
	os := strings.Split(string(platform), "/")[0]

	fmt.Println("Set build add for arch:", arch)

	oses = []string{os}
	arches = []string{arch}

	outputArch := arch
	if outputArch == "amd64" {
		outputArch = "x86_64"
	}

	// only change notorize if we are on darwin
	if os == "darwin" {
		filename := strings.Replace("kapsule_%%VERSION%%_darwin_%%ARCH%%.zip", "%%ARCH%%", outputArch, 1)
		notorize = []Archive{
			{Path: filename, Type: "zip", Output: filename},
		}

		archives = []Archive{
			{Path: fmt.Sprintf("/build/darwin/%s/kapsule.exe", arch), Type: "zip", Output: filename},
		}
	}

	if os == "linux" {
		filename := strings.Replace("kapsule_%%VERSION%%_darwin_%%ARCH%%.tar.gz", "%%ARCH%%", outputArch, 1)
		filenameDeb := strings.Replace("kapsule_%%VERSION%%_darwin_%%ARCH%%.deb", "%%ARCH%%", outputArch, 1)

		archives = []Archive{
			{Path: fmt.Sprintf("/build/linux/%s/kapsule", arch), Type: "targz", Output: filename},
			{Path: fmt.Sprintf("/pkg/linux/%s/kapsule.deb", arch), Type: "copy", Output: filenameDeb},
		}
	}
}
