package logging

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry/cf-smoke-tests/smoke"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)


var _ = Describe("Loggregator:", func() {
	var testConfig = smoke.GetConfig()
	var useExistingApp = testConfig.LoggingApp != ""
	var appName string
	var manifestPath string

	Describe("cf logs", func() {
		AfterEach(func() {
			defer func() {
				if testConfig.Cleanup && !useExistingApp {
					Expect(cf.Cf("delete", appName, "-f", "-r").Wait(testConfig.GetDefaultTimeout())).To(Exit(0))
				}
			}()
			smoke.AppReport(appName, testConfig.GetDefaultTimeout())
		})

		Context("linux", func() {
			BeforeEach(func() {
				if !useExistingApp {
					appName = generator.PrefixedRandomName("SMOKES", "APP")
					manifestPath = CreateManifestWithRoute(appName, testConfig.AppsDomain)
					Expect(cf.Cf("push", appName,
						"-b", testConfig.BinaryBuildpack,
						"-m", "30M",
						"-k", "16M",
						"-p", smoke.SimpleBinaryAppBitsPath,
						"-f", manifestPath).Wait(testConfig.GetPushTimeout())).To(Exit(0))
				} else {
					appName = testConfig.LoggingApp
				}
			})

			It("can see app messages in the logs", func() {
				Eventually(func() *Session {
					appLogsSession := smoke.Logs(testConfig.UseLogCache, appName)
					Expect(appLogsSession.Wait(testConfig.GetDefaultTimeout())).To(Exit(0))

					return appLogsSession
				}, testConfig.GetDefaultTimeout()*5).Should(Say(`\[(App|APP).*/0\]`))
			})
		})

		Context("windows", func() {
			BeforeEach(func() {
				smoke.SkipIfNotWindows(testConfig)
				appName = generator.PrefixedRandomName("SMOKES", "APP")
				manifestPath = CreateManifestWithRoute(appName, testConfig.AppsDomain)
				Expect(cf.Cf("push", appName,
					"-p", smoke.SimpleDotnetAppBitsPath,
					"-s", testConfig.GetWindowsStack(),
					"-f", manifestPath,
					"-b", "hwc_buildpack").Wait(testConfig.GetPushTimeout())).To(Exit(0))
			})

			It("can see app messages in the logs", func() {
				Eventually(func() *Session {
					appLogsSession := cf.Cf("logs", "--recent", appName)
					Expect(appLogsSession.Wait(testConfig.GetDefaultTimeout())).To(Exit(0))
					return appLogsSession
				}, testConfig.GetDefaultTimeout()*5).Should(Say(`\[(App|APP).*/0\]`))
			})
		})
	})
})

func CreateManifestWithRoute(name string, domain string) string {
	file, err := ioutil.TempFile(os.TempDir(), "iso-segment-manifest-*.yml")
	Expect(err).NotTo(HaveOccurred())

	filePath := file.Name()

	_, err = file.Write([]byte(fmt.Sprintf("---\n" +
		"applications:\n" +
		"- name: %s\n" +
		"  routes:\n" +
		"  - route: %s.%s",
		name, name, domain)))
	Expect(err).NotTo(HaveOccurred())

	err = file.Close()
	Expect(err).NotTo(HaveOccurred())

	return filePath
}
