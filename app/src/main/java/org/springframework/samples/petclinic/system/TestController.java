package org.springframework.samples.petclinic.system;
import org.springframework.dao.DataAccessException;
import org.springframework.dao.QueryTimeoutException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.samples.petclinic.owner.*;
import org.springframework.web.bind.annotation.*;

@RestController
public class TestController {
    private static final int TEST_OWNER_ID = 1;
    private static final int TEST_PET_ID = 1;

    private final OwnerRepository owners;
    private final PetRepository pets;

    public TestController(OwnerRepository clinicService, PetRepository pets) {
        this.owners = clinicService;
        this.pets = pets;
    }

    /**
     * 사전 과제 요구사항 테스트 용도의 REST API
     * !! Production 환경에서 이 API를 오픈하면 안되며, 실제 환경에선 이 API의 접근을 막아야 하지만 시간 관계상 주석으로만...
     * 방법 1 - 프로덕션에서도 이 API를 관리자가 테스트 용도로 필요하다면 액세스 토큰 값을 같이 전달하여 해당 API를 액세스 할 수 있는 권한을 설정
     * 방법 2 - 프로덕션에서 해당 API에 대한 접근 제한을 내부망 기반으로 관리
     * 방법 3 - 프로덕션 컴파일시에 해당 API 제거
     * @param timeSec 해당 값만큼 대기 한다.
     * @return 대기 후, StatusCode 200 반환
     * @throws InterruptedException
     */
    @GetMapping("/pause/{timeSec}")
    public String pause(@PathVariable int timeSec) throws InterruptedException {
        Thread.sleep(timeSec * 1000);
        return "Process finished";
    }

    /**
     * Health Check 용도의 REST API.
     * DB 액세스에 오류가 발생하면 전체 서비스가 불가능 하다고 판단하는 극단적인 케이스의 상태 체크이다.
     * @return 성공시 200 "OK" 반환
     */
    @GetMapping("/health")
    public String health() {
        this.owners.findById(TEST_OWNER_ID);
        this.pets.findById(TEST_PET_ID);

        return "OK";
    }
}
