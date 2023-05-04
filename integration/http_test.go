package integration_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("HTTP", func() {
	var (
		publisherPath string

		srv  *httptest.Server
		port string
	)

	BeforeEach(func() {
		var err error
		publisherPath, err = gexec.Build("go.chensl.me/minitunnel/cmd/mt")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(gexec.CleanupBuildArtifacts)

		srv = httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Hello World! %s", time.Now())
			}),
		)
		DeferCleanup(srv.Close)

		port = srv.URL[strings.LastIndexByte(srv.URL, ':')+1:]
	})

	It("hello web HTTP server", func(ctx SpecContext) {
		By("Setup mt server")
		cmd1 := exec.CommandContext(ctx, publisherPath, "server")
		session1, err := gexec.Start(cmd1, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		defer session1.Kill()

		By("Setup mt client")
		cmd2 := exec.CommandContext(
			ctx,
			publisherPath,
			"local",
			"-t",
			"localhost",
			port,
		)
		session2, err := gexec.Start(cmd2, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		defer session2.Kill()

		By("Wait for listen success")
		Eventually(ctx, func() string {
			return string(session2.Err.Contents())
		}).Should(ContainSubstring("listening at"))

		re := regexp.MustCompile(`tcp://(.*)`)
		match := re.FindSubmatch(session2.Err.Contents())
		Expect(len(match)).Should(Equal(2))

		By("send request")
		resp, err := http.Get("http://" + string(match[1]))
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	}, SpecTimeout(10*time.Second))
})
