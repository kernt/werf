// +build integration

package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/flant/werf/integration/utils"
	utilsDocker "github.com/flant/werf/integration/utils/docker"
)

var _ = Describe("Advanced build/First application", func() {
	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	for _, builder := range []string{"shell", "ansible"} {
		boundedBuilder := builder

		It(fmt.Sprintf("%s application should be built, checked and published", boundedBuilder), func() {
			utils.RunSucceedCommand(
				".",
				"git",
				"clone", "https://github.com/symfony/symfony-demo.git", testDirPath,
			)

			utils.CopyIn(fixturePath("first_application", boundedBuilder), testDirPath)

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"build", "-s", ":local",
			)

			containerHostPort := utils.GetFreeTCPHostPort()
			containerName := fmt.Sprintf("symfony_demo_%s_%s", boundedBuilder, utils.GetRandomString(10))
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p %d:8000 --name %s", containerHostPort, containerName), "--", "/app/start.sh",
			)
			defer func() { utilsDocker.ContainerStopAndRemove(containerName) }()

			url := fmt.Sprintf("http://localhost:%d", containerHostPort)
			waitTillHostReadyAndCheckResponseBody(
				url,
				240,
				"Symfony Demo application",
			)

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"publish", "-s", ":local", "-i", registryProjectRepository, "--tag-custom", "test",
			)
		})
	}
})
