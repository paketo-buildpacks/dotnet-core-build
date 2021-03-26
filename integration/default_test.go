package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		pack       occam.Pack
		docker     occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building an app", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
			source    string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("should build a working OCI image for an app that contains a directory with the same name as the app", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "match_dir_and_app_name"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(
					icuBuildpack,
					dotnetCoreRuntimeBuildpack,
					dotnetCoreSDKBuildpack,
					buildpack,
					dotnetExecuteBuildpack,
				).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"    WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Publish Buildpack v2.0.0",
				"    Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information.",
				"  Executing build process",
				MatchRegexp(`    Running 'dotnet publish \/workspace\/console --configuration Release --runtime ubuntu\.18\.04-x64 --self-contained false --output \/tmp\/dotnet-publish-output\d+'`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Removing source code",
				"",
			))

			container, err = docker.Container.Run.
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(ContainSubstring("Hello World!"))
		})

		context("when custom project path is set via $BP_DOTNET_PROJECT_PATH", func() {
			it("should build a working OCI image for an app", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "match_dir_and_app_name"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						icuBuildpack,
						dotnetCoreRuntimeBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
						dotnetExecuteBuildpack,
					).
					WithEnv(map[string]string{"BP_DOTNET_PROJECT_PATH": "console"}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
					"  Executing build process",
					MatchRegexp(`    Running 'dotnet publish \/workspace\/console --configuration Release --runtime ubuntu\.18\.04-x64 --self-contained false --output \/tmp\/dotnet-publish-output\d+'`),
					MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
					"",
					"  Removing source code",
					"",
				))

				Expect(logs).ToNot(ContainSubstring("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Publish Buildpack v1.0.0"))

				container, err = docker.Container.Run.
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(ContainSubstring("Hello World!"))
			})
		})
	})
}
