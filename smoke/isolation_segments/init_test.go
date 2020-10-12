package isolation_segments

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cf-smoke-tests/smoke"
)

var (
	testConfig *smoke.Config
	testSetup  *workflowhelpers.ReproducibleTestSuiteSetup
)

func TestSmokeTests(t *testing.T) {
	RegisterFailHandler(Fail)

	testConfig = smoke.GetConfig()
	testSetup = workflowhelpers.NewSmokeTestSuiteSetup(testConfig)

	SynchronizedBeforeSuite(
		func() []byte {
			return nil
		},
		func(data []byte) {
			testSetup.Setup()
		},
	)

	SynchronizedAfterSuite(
		func() {
			testSetup.Teardown()
		},
		func() {},
	)
	rs := []Reporter{}

	if testConfig.ArtifactsDirectory != "" {
		os.Setenv("CF_TRACE", traceLogFilePath(testConfig))
		rs = append(rs, reporters.NewJUnitReporter(jUnitReportFilePath(testConfig)))
	}

	if testConfig.Reporter == "TeamCity" {
		rs = append(rs, reporters.NewTeamCityReporter(GinkgoWriter))
	}

	RunSpecsWithDefaultAndCustomReporters(t, "CF-Isolation-Segment-Smoke-Tests", rs)
}

func traceLogFilePath(testConfig *smoke.Config) string {
	return filepath.Join(testConfig.ArtifactsDirectory, fmt.Sprintf("CF-TRACE-%s-%d.txt", testConfig.SuiteName, ginkgoNode()))
}

func jUnitReportFilePath(testConfig *smoke.Config) string {
	return filepath.Join(testConfig.ArtifactsDirectory, fmt.Sprintf("junit-%s-%d.xml", testConfig.SuiteName, ginkgoNode()))
}

func ginkgoNode() int {
	return ginkgoconfig.GinkgoConfig.ParallelNode
}
