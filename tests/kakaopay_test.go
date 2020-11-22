package tests

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func Setup() {
	exec.Command("kubectl", "scale", "deployment", "--replicas", "1", "kakaopay-app").Run()
	exec.Command("kubectl", "scale", "deployment", "--replicas", "1", "mysql").Run()

	fmt.Print("테스트 환경을 구성하고 완료될때까지 기다립니다.\n")

	WaitUntilSuccess("http://localhost/")
}

func WaitUntilSuccess(url string) {
	for {
		resp, _ := http.Get(url)
		if resp != nil {
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				break
			}
		}

		time.Sleep(time.Second * 1)
	}
}

func Request(url string, statusCodeChan chan int) {
	resp, _ := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}

	statusCodeChan <- resp.StatusCode
}

// -Problem-
// 정상 동작 여부를 반환하는 api를 구현하며, 10초에 한번 체크하도록 한다. 3번 연속 체크에 실패하면 어플리케이션은 restart 된다.
//
// -Solve-
// k8s의 liveness probe를 사용하여 해결
//
// -Test senario-
// 이 테스트 시나리오는 앱 서버가 database에 연결이 불가능 한 상황을 비정상 상황으로 판단한다.
func Test_Problem_3(t *testing.T) {
	Setup()

	// Database pod를 내린다.
	fmt.Print("Database pod를 내립니다.\n")
	exec.Command("kubectl", "scale", "deployment", "--replicas", "0", "mysql").Run()

	// liveness probe에 따라 health check가 실패하여 모든 앱 pod는 계속 restart를 진행하게 될 것이다.
	// 이에 따라 요청이 실패가 와야 함
	numberOfFailure := 0
	for {
		resp, _ := http.Get("http://localhost/health")
		if resp != nil {
			if resp.StatusCode != 200 {
				fmt.Printf("요청 실패 횟수: %d\n", numberOfFailure)
				numberOfFailure++
			} else {
				numberOfFailure = 0
			}

			resp.Body.Close()

			if numberOfFailure >= 5 {
				break
			}
		}

		time.Sleep(time.Second * 1)
	}

	// Database pod를 다시 올린다.
	fmt.Print("Database pod를 다시 올립니다.\n")
	exec.Command("kubectl", "scale", "deployment", "--replicas", "1", "mysql").Run()

	// 요청이 성공 할 때까지 대기해본다.
	fmt.Print("요청이 성공할때까지 기다립니다.\n")
	WaitUntilSuccess("http://localhost/health")
}

// -Problem-
// 종료 시 30초 이내에 프로세스가 종료되지 않으면 SIGKILL로 강제 종료 시킨다.
//
// -Solve-
// k8s는 SIGTERM으로 종료를 시도하고 Terminating 상태로 들어가는데 컨테이너의 프로세스가
// 정상 종료 되지 않는다면 terminationGracePeriodSeconds 값에 따라(디폴트 30초) SIGKILL로 강제 종료 시킨다.
//
// -Test senario-
// 서버 측에서 처리하는데 40초가 걸리는 REST API를 콜한 뒤, pod를 모두 내린다.
func Test_Problem_4(t *testing.T) {
	Setup()

	ch := make(chan int)
	go Request("http://localhost/pause/40", ch)

	fmt.Print("요청 중 모든 app pod를 내립니다.\n")
	exec.Command("kubectl", "scale", "deployment", "--replicas", "0", "kakaopay-app").Run()

	if <-ch == 200 {
		t.Error()
	}
}

// -Problem-
// 배포 시와 scale in/out 시 유실되는 트래픽이 없어야 한다.
//
// -Solve-
// 웹 어플리케이션에서 graceful shutdown을 기능을 사용한다.
//
// -Test senario-
// 서버 측에서 처리하는데 10초(실제 서비스시에서는 극한 상황)가 걸리는 REST API를 콜한 뒤, pod를 모두 내린다.
func Test_Problem_5(t *testing.T) {
	Setup()

	ch := make(chan int)
	go Request("http://localhost/pause/10", ch)

	fmt.Print("요청 중 모든 app pod를 내립니다.\n")
	exec.Command("kubectl", "scale", "deployment", "--replicas", "0", "kakaopay-app").Run()

	if <-ch != 200 {
		t.Error()
	}
}

func TearDown() {
	exec.Command("kubectl", "scale", "deployment", "--replicas", "3", "kakaopay-app").Run()
	exec.Command("kubectl", "scale", "deployment", "--replicas", "1", "mysql").Run()
}

func TestMain(m *testing.M) {
	retCode := m.Run()

	TearDown()

	os.Exit(retCode)
}
