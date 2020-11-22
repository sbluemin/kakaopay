# 테스트 환경

- OS: Ubuntu 20.04
- Java: 14.0.2
- Docker: 19.03.8
- k8s: microk8s

# 환경 세팅 (New VM)

**IMPORTANT: 아래 커맨드 마지막에 reboot을 하는 커맨드가 들어가 있으므로 테스트 하시는 머신에 따라 주의가 필요 합니다.**

```bash
sudo apt update &&
sudo apt install -y openjdk-14-jdk docker.io golang &&
sudo snap install microk8s --classic &&
sudo microk8s.start &&
sudo microk8s.enable dns &&
sudo microk8s.enable ingress &&
sudo snap alias microk8s.kubectl kubectl &&
sudo iptables -P FORWARD ACCEPT &&
sudo usermod -aG microk8s $USER &&
sudo usermod -aG docker $USER &&
sudo reboot now
```

# 참고

- ingress-nginx-controller를 커스텀 하지 않고 microk8s의 ingress 기능을 사용 하였습니다.  
  따라서 위의 세팅으로 진행하지 않고 이미 구축된 환경으로 테스트가 진행 된다면 **ingress-nginx-controller**가 설치 되어 있다고 가정합니다.  
  (microk8s의 경우 `microk8s.enable ingress` 로 활성화)

# k8s 앱 환경 세팅

## 빌드

코드 베이스의 **build.sh** 파일이 앱 이미지 빌드를 하여 로컬 레지스트리에 저장합니다.  

> 과제의 외부 의존성을 줄이기 위해 WAN에서 pull 하지 않고 로컬 레지스트리에서만 사용 하도록 하려 했지만 k8s 구성 환경에 따라 차이가 발생하여 테스트 환경이 어떻게 되는지를 몰라 k8s의 매니페스트에서는 저의 개인 docker hub에서 이미지를 pull 하도록 설정 되어 있습니다. 해당 빌드 스크립트는 요구 사항 1번을 충족 하기 위함입니다.

## 구성

코드 베이스의 **setup.sh** 파일이 v1 이미지로 앱 환경을 세팅합니다.


> Docker hub에는 같은 레이어의 v1과 v2가 단순히 태그로만 나뉘어져 존재합니다.  
이는 롤링 업데이트 테스트를 위함입니다.

# 테스트

- [http://localhost](http://localhost) 접속 시에 spring-petclinic-jdbc 웹 사이트로 접속 할 수 있습니다.
- 수동 테스트가 필요한 일부 요구 사항은 코드 베이스의 **tests/kakaopay_test.go**로 구현이 되어 있습니다.  
해당 디렉토리에서 `go test -v`로 테스트를 진행 할 수 있습니다.
- 유닛 테스트로만 충족하기 힘든 **트래픽 유실** 에 대한 정확한 테스트를 위하여 **tests** 디렉토리에 **load_test_deploy.sh** 가 구현되어 있으며 이를 통해 부하 기반으로 트래픽 유실 케이스에 대한 테스트가 가능합니다.


> 모든 테스트 이전에 반드시 최초 **k8s 앱 환경 세팅 → 구성** 작업이 선결 되어야 합니다.

> 유닛 테스트의 경우 출력이 따로 없지만 시간이 좀 오래 걸리는 테스트이므로 기다리시면 됩니다.  
또한, 유닛 테스트 도중 app pod를 건드리면 테스트의 정합성이 떨어질 수 있으므로 주의 해야 합니다.  
(대략 2분 정도 걸립니다.)

> 테스트를 진행하면서 다른 쉘에서 pod의 모니터링을 같이 한다면 더욱 편합니다.  
> watch -n -1 kubectl get pod

# 요구 사항 정리

- gradle을 사용하여 어플리케이션과 도커이미지를 빌드한다.  
→ maven 기반으로 되어 있던 프로젝트를 gradle로 마이그레이션 하였습니다.  

- 어플리케이션의 log는 host의 /logs 디렉토리에 적재되도록 한다.  
→ /logs 디렉토리를 미리 생성하고 소유자와 소유그룹을 1000:1000 설정 한 뒤, 700 권한으로 변경 하여 hostPath로 마운트 하였습니다.  

- 정상 동작 여부를 반환하는 api를 구현하며, 10초에 한번 체크하도록 한다. 3번 연속 체크에 실패하면 어플리케이션은 restart 된다.  
→ REST API 컨트롤러를 만들어 TestController에 /health API를 제작 후, app-deployment.yaml에 liveness probe를 적용 하였습니다.  

- 종료 시 30초 이내에 프로세스가 종료되지 않으면 SIGKILL로 강제 종료 시킨다.  
→ k8s는 SIGTERM으로 종료 요청 후에 terminationGracePeriodSeconds 값이 지나면 SIGKILL을 진행 합니다. 따라서 app-deployment.yaml에 해당 스펙을 명시하여 해결 하였습니다.  

- 배포 시와 scale in/out 시 유실되는 트래픽이 없어야 한다.  
→ 웹 어플리케이션에서 Graceful Shutdown 기능을 사용하여 해결 하였습니다.  

- 어플리케이션 프로세스는 root 계정이 아닌 uid:1000으로 실행한다.  
→ Spring Boot 2.3.0-RELEASE 부터 지원하는 자체 도커 이미지 제작시에는 기본적으로 cnb 유저로 uid:1000으로 실행을 하게 되어 있습니다.  

- DB도 kubernetes에서 실행하며 재 실행 시에도 변경된 데이터는 유실되지 않도록 설정한다.  
→ hostPath를 사용하여 Pod에 마운트를 한 뒤, 해결 하였습니다.  

- nginx-ingress-controller를 통해 어플리케이션에 접속이 가능하다.  
→ 과제 수행 환경인 microk8s의 ingress를 활성화 하여 해결 하였습니다.  

- namespace는 default를 사용한다.  
→ 앱과 DB는 기본적으로 default를 사용 하도록 하였습니다.

# 아쉬운 점
1. 유닛 테스트가 완벽한 블랙박스 테스트이다.  
golang에 k8s 라이브러리가 있는데 이것을 활용하면 mock 기반 화이트박스 테스트가 가능 할 것 같다.  

2. spring boot 프레임워크에 대한 이해도가 낮아 빌드 환경을 구성하는데 꽤 많은 시간을 들였다.  
특히, less 파일을 컴파일하고 리소스들을 bundling 해주는 wro4j가 gradle에 정상적으로 적용이 되지 않아 maven에서 컴파일한 .css 파일을 그대로 포함하여 배포하는 구조로 되어 있는것이 현재 결함이다.

1. liveness probe에 대한 유닛 테스트를 최대한 best practice에 가깝게 해보려 했는데 완벽한 정답은 없는 것 같다.   
현재 작성 된 테스트 시나리오도 개인적으로도 이뻐 보이지는 않는다.