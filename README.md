# Kubernetes Job 테스트용 어플리케이션

[![GitHub - tryoo0607/job-test](https://img.shields.io/badge/GitHub-tryoo0607%2Fjob--test-181717?logo=github&logoColor=white&style=flat)](https://github.com/tryoo0607/job-test)
[![Docker Hub - tryoo0607/job-test](https://img.shields.io/badge/Docker%20Hub-tryoo0607%2Fjob--test-2496ED?logo=docker&logoColor=white&style=flat)](https://hub.docker.com/r/tryoo0607/job-test)

<br/>

이 프로젝트는 **Kubernetes Job 환경에서 다양한 작업 분산 처리 패턴**을 실습하기 위한 예제 애플리케이션입니다.  
텍스트 파일을 대문자로 변환하여 출력 파일로 저장하며, 다양한 실행 모드를 통해 **IndexedJob**, **Fixed Worker Pool**, **Queue 기반 분산 처리**, **Pod 간 통신(Peer)** 등을 실습할 수 있습니다.

<br/>

## ✨ 주요 기능

- ✅ **텍스트 파일 대문자 변환**
- ✅ **Retry with Exponential Backoff**
- ✅ **다양한 실행 모드 지원**
  - `indexed`: Job Index 기반 입력 처리 (K8s IndexedJob 사용)
  - `fixed`: 고정된 아이템 리스트를 워커 풀로 병렬 처리
  - `queue`: Redis Queue 기반 분산 처리
  - `peer`: 인접 Pod와 통신 (Indexed 기반 Peer 구조)

---

## 🧱 Docker 빌드

```bash
docker build -t job-test -f docker/Dockerfile .
```

- ✅ 빌드는 `/cmd/entrypoint/main.go` 기준으로 수행됩니다.
- Dockerfile에는 `ENTRYPOINT`가 정의되어 있지 않으므로, **Kubernetes에서 `command:`를 통해 실행 모드 지정**이 필요합니다.

---

## 🐳 로컬 실행 예시

```bash
# Indexed 모드
docker run --rm -e JOB_COMPLETION_INDEX=0 -v $(pwd)/assets/inputs:/data/inputs -v $(pwd)/outputs:/data/outputs job-test --mode=indexed

# Fixed 모드
docker run --rm -v $(pwd)/assets/inputs:/data/inputs -v $(pwd)/outputs:/data/outputs job-test --mode=fixed --items="/data/inputs/input-0.txt,/data/inputs/input-1.txt"

# Queue 모드 (Redis 필요)
docker run --rm -e QUEUE_URL=redis://localhost:6379 -v $(pwd)/outputs:/data/outputs job-test --mode=queue

# Peer 모드 (예: index 0 → index 1로 요청)
docker run --rm -e JOB_COMPLETION_INDEX=0 job-test --mode=peer --total-pods=3 --subdomain=myjob.default.svc.cluster.local
```

---

## ☸️ Kubernetes 예시

### IndexedJob (기본)

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: job-test-indexed
spec:
  completions: 3
  parallelism: 3
  completionMode: Indexed
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: job
        image: tryoo0607/job-test:latest
        command: ["/job-test", "--mode=indexed"]
        env:
        - name: JOB_COMPLETION_INDEX
          valueFrom:
            fieldRef:
              fieldPath: metadata.annotations['batch.kubernetes.io/job-completion-index']
        volumeMounts:
        - name: input-vol
          mountPath: /data/inputs
        - name: output-vol
          mountPath: /data/outputs
      volumes:
      - name: input-vol
        hostPath:
          path: /your/local/path/assets/inputs
      - name: output-vol
        hostPath:
          path: /your/local/path/outputs
```

---

## 📁 입력 / 출력 예시

**입력 파일:**

```txt
# /data/inputs/input-0.txt
hello world

# /data/inputs/input-1.txt
kubernetes job test

# /data/inputs/input-2.txt
index based processing
```

**출력 예시 (Indexed 기준):**

```txt
# /data/outputs/output-0.txt
HELLO WORLD

# /data/outputs/output-1.txt
KUBERNETES JOB TEST

# /data/outputs/output-2.txt
INDEX BASED PROCESSING
```

---

## ⚙️ Config / CLI 옵션

| 옵션 또는 ENV               | 설명                                            |
|----------------------------|-------------------------------------------------|
| `--mode`                   | 실행 모드: `indexed`, `fixed`, `queue`, `peer` |
| `--max-concurrency`        | 최대 병렬 처리 수 (default: 1)                 |
| `--retry-max`              | 재시도 횟수 (기본 0)                           |
| `--retry-backoff`          | 재시도 대기 시간 (예: `2s`)                    |
| `--items`                  | 처리할 파일 경로 리스트 (쉼표 구분)            |
| `--items-file`             | 처리 대상이 담긴 파일 경로                    |
| `--queue-url`              | Redis URL (`redis://localhost:6379`)           |
| `--queue-key`              | Redis 작업 큐 이름 (default: tasks)           |
| `--input-dir`              | 입력 디렉토리 경로 (default: `/data/inputs`)  |
| `--output-dir`             | 출력 디렉토리 경로 (default: `/data/outputs`) |
| `--subdomain`              | Peer 모드에서 사용할 서브도메인 이름           |
| `--total-pods`             | Peer 구조에서 총 Pod 수                         |
| `JOB_COMPLETION_INDEX`     | Indexed Job Index 값 (env 전용)                |

---

## ✅ 참고

- Peer 모드는 실제 통신 로직을 붙이거나, 단순 로그 출력으로 대체할 수 있습니다.
- `queue` 모드 사용 시, Redis 서버가 별도로 필요합니다.
- 입력 파일은 컨테이너 빌드 시 포함하거나, `hostPath`, PVC, ConfigMap 등으로 주입할 수 있습니다.