// +build integ
// Copyright Istio Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy

import (
	"errors"
	"io/ioutil"
	"testing"
	"time"
	"path/filepath"

	"istio.io/istio/pkg/config/protocol"
	"istio.io/istio/pkg/test/echo/common/response"
	"istio.io/istio/pkg/test/framework"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/components/echo/echoboot"
	"istio.io/istio/pkg/test/framework/components/istio"
	"istio.io/istio/pkg/test/framework/components/istio/ingress"
	"istio.io/istio/pkg/test/framework/components/namespace"
	"istio.io/istio/pkg/test/framework/label"
	"istio.io/istio/pkg/test/framework/resource"
	"istio.io/istio/pkg/test/kube"
	"istio.io/istio/pkg/test/util/retry"
	"istio.io/istio/pkg/test/util/tmpl"
	"istio.io/istio/pkg/test/env"
)

var (
	ist         istio.Instance
	echoNsInst  namespace.Instance
	ratelimitNs namespace.Instance
	ing         ingress.Instance
	srv         echo.Instance
	clt         echo.Instance
)

func TestRateLimiting(t *testing.T) {
	framework.
		NewTest(t).
		Features("traffic.ratelimit.envoy").
		Run(func(ctx framework.TestContext) {
			setupEnvoyFilter(ctx, "testdata/enable_envoy_ratelimit.yaml")
			sendTrafficAndCheckIfRatelimited(t)
		})
}

func TestLocalRateLimiting(t *testing.T) {
	framework.
		NewTest(t).
		Features("traffic.ratelimit.envoy").
		Run(func(ctx framework.TestContext) {
			setupEnvoyFilter(ctx, "testdata/enable_envoy_local_ratelimit.yaml")

			sendTrafficAndCheckIfRatelimited(t)
		})
}

func TestLocalRouteSpecificRateLimiting(t *testing.T) {
	framework.
		NewTest(t).
		Features("traffic.ratelimit.envoy").
		Run(func(ctx framework.TestContext) {
			setupEnvoyFilter(ctx, "testdata/enable_envoy_local_ratelimit_per_route.yaml")

			sendTrafficAndCheckIfRatelimited(t)
		})
}

func TestMain(m *testing.M) {
	framework.
		NewSuite(m).
		RequireSingleCluster().
		Label(label.CustomSetup).
		Setup(istio.Setup(&ist, nil)).
		Setup(testSetup).
		Run()
}

func testSetup(ctx resource.Context) (err error) {
	echoNsInst, err = namespace.New(ctx, namespace.Config{
		Prefix: "istio-echo",
		Inject: true,
	})
	if err != nil {
		return
	}

	_, err = echoboot.NewBuilder(ctx).
		With(&clt, echo.Config{
			Service:   "clt",
			Namespace: echoNsInst,
		}).
		With(&srv, echo.Config{
			Service:   "srv",
			Namespace: echoNsInst,
			Ports: []echo.Port{
				{
					Name:     "http",
					Protocol: protocol.HTTP,
					// We use a port > 1024 to not require root
					InstancePort: 8888,
				},
			},
		}).
		Build()
	if err != nil {
		return
	}

	ing = ist.IngressFor(ctx.Clusters().Default())

	ratelimitNs, err = namespace.New(ctx, namespace.Config{
		Prefix: "istio-ratelimit",
	})
	if err != nil {
		return
	}

	yamlContentCM, err := ioutil.ReadFile("testdata/rate-limit-configmap.yaml")
	if err != nil {
		return
	}

	err = ctx.Config().ApplyYAML(ratelimitNs.Name(),
		string(yamlContentCM),
	)
	if err != nil {
		return
	}

	yamlContent, err := ioutil.ReadFile(filepath.Join(env.IstioSrc, "samples/ratelimit/ratelimit.yaml"))
	if err != nil {
		return
	}

	err = ctx.Config().ApplyYAML(ratelimitNs.Name(),
		string(yamlContent),
	)
	if err != nil {
		return
	}

	// Wait for redis and ratelimit service to be up.
	fetchFn := kube.NewPodFetch(ctx.Clusters().Default(), ratelimitNs.Name(), "app=redis")
	if _, err = kube.WaitUntilPodsAreReady(fetchFn); err != nil {
		return
	}
	fetchFn = kube.NewPodFetch(ctx.Clusters().Default(), ratelimitNs.Name(), "app=ratelimit")
	if _, err = kube.WaitUntilPodsAreReady(fetchFn); err != nil {
		return
	}

	return nil
}

func setupEnvoyFilter(ctx framework.TestContext, file string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		ctx.Fatal(err)
	}

	con, err := tmpl.Evaluate(string(content), map[string]interface{}{
		"EchoNamespace":      echoNsInst.Name(),
		"RateLimitNamespace": ratelimitNs.Name(),
	})
	if err != nil {
		ctx.Fatal(err)
	}

	err = ctx.Config().ApplyYAML(ist.Settings().SystemNamespace, con)
	if err != nil {
		ctx.Fatal(err)
	}
}

func sendTrafficAndCheckIfRatelimited(t *testing.T) {
	t.Helper()
	retry.UntilSuccessOrFail(t, func() error {
		t.Logf("Sending 5 requests...")
		httpOpts := echo.CallOptions{
			Target:   srv,
			PortName: "http",
			Count:    5,
		}
		received409 := false
		if parsedResponse, err := clt.Call(httpOpts); err == nil {
			for _, resp := range parsedResponse {
				if response.StatusCodeTooManyRequests == resp.Code {
					received409 = true
					break
				}
			}
		}
		if !received409 {
			return errors.New("no request received StatusTooManyRequest error")
		}
		return nil
	}, retry.Delay(10*time.Second), retry.Timeout(60*time.Second))
}
