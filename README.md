# Kubernetes Job 테스트용 어플리케이션

[![GitHub - tryoo0607/job-test](https://img.shields.io/badge/GitHub-tryoo0607%2Fjob--test-181717?logo=github&logoColor=white&style=flat)](https://github.com/tryoo0607/job-test)
[![Docker Hub - tryoo0607/job-test](https://img.shields.io/badge/Docker%20Hub-tryoo0607%2Fjob--test-2496ED?logo=docker&logoColor=white&style=flat)](https://hub.docker.com/r/tryoo0607/job-test)

<br/>

이 프로젝트는 **Kubernetes Job 환경에서 다양한 작업 분산 처리 패턴**을 실습하기 위한 예제 애플리케이션입니다.  
텍스트 파일을 대문자로 변환하여 출력 파일로 저장하며, 다양한 실행 모드를 통해 **IndexedJob**, **Fixed Worker Pool**, **Queue 기반 분산 처리**, **Peer 간 통신** 등을 실습할 수 있습니다.

<br/>
<br/>

## ✨ 주요 기능

- **텍스트 파일 대문자 변환**
  - 주어진 입력 파일의 내용을 읽어 대문자로 변환하여 출력 디렉토리에 저장

- **Retry with Exponential Backoff**
  - 작업 실패 시 재시도 로직 내장

- **다양한 실행 모드 지원**
  - `indexed`: Job Index 기반 입력 처리 (K8s IndexedJob 사용)
  - `fixed`: 고정된 아이템 리스트를 워커 풀로 병렬 처리
  - `queue`: Redis Queue를 통해 작업 분산 처리
  - `indexed-peer`: 인접 Pod 간 HTTP 통신 수행 (Peer 테스트용)

---

## 🧱 실행 모드 설명

| 모드            | 설명                                                                 |
|-----------------|----------------------------------------------------------------------|
| `indexed`       | `$JOB_COMPLETION_INDEX`를 기반으로 input-<N>.txt 하나만 처리합니다. |
| `fixed`         | 입력 리스트(또는 파일)를 읽어 여러 아이템을 고루 처리합니다.         |
| `queue`         | Redis에서 작업을 꺼내 비동기 처리합니다.                              |
| `indexed-peer`  | 본인의 Index를 기준으로 다음 Peer에 HTTP 요청을 보냅니다.            |

---

## 🏗️ Docker 빌드

```bash
docker build -t job-test .
```

---

## 🐳 로컬 실행 예시

```bash
# Indexed 모드
docker run --rm -e JOB_COMPLETION_INDEX=0 -v $(pwd)/assets/inputs:/data/inputs -v $(pwd)/outputs:/data/outputs job-test --mode=indexed

# Fixed 모드
docker run --rm -v $(pwd)/assets/inputs:/data/inputs -v $(pwd)/outputs:/data/outputs job-test --mode=fixed --items="/data/inputs/input-0.txt,/data/inputs/input-1.txt"

# Queue 모드 (Redis 필요)
docker run --rm -e QUEUE_URL=redis://localhost:6379 -v $(pwd)/outputs:/data/outputs job-test --mode=queue
```

---

## ☸️ Kubernetes 예시

### IndexedJob 예시

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
        image: tryoo0607/job-test
        args: ["--mode=indexed"]
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
          path: /path/to/assets/inputs
      - name: output-vol
        hostPath:
          path: /path/to/outputs
```

---

## 📁 입력 파일 예시

```txt
# /data/inputs/input-0.txt
hello world

# /data/inputs/input-1.txt
kubernetes job test

# /data/inputs/input-2.txt
index based processing
```

출력 결과는 `/data/outputs/output-<index>.txt`로 저장됩니다:

```
HELLO WORLD
KUBERNETES JOB TEST
INDEX BASED PROCESSING
```

---

## 📦 Config 옵션

| 환경 변수 / 플래그       | 설명                               |
|--------------------------|------------------------------------|
| `--mode`                 | 실행 모드 (`indexed`, `fixed` 등) |
| `--max-concurrency`      | 병렬 처리 수 (default: 1)         |
| `--retry-max`            | 최대 재시도 횟수                  |
| `--retry-backoff`        | 재시도 초기 대기 시간             |
| `--items`                | 처리할 파일 목록 (comma-separated)|
| `--items-file`           | 처리할 파일 경로 목록이 담긴 파일 |
| `--queue-url`            | Redis 주소 (`redis://...`)        |
| `--queue-key`            | Redis 작업 큐 이름 (default: tasks) |
| `--input-dir`            | 입력 디렉토리 경로 (default: /data/inputs) |
| `--output-dir`           | 출력 디렉토리 경로 (default: /data/outputs) |
| `JOB_COMPLETION_INDEX`  | Indexed Job용 Index (env 전용)    |

---

## 🔗 관련

- [Kubernetes Indexed Jobs 설명](https://kubernetes.io/docs/concepts/workloads/controllers/job/#indexed-job)
- [SIGTERM 처리 실습 예제](https://github.com/tryoo0607/pod-lifecycle-test)

---

## 🧑‍💻 Author

- GitHub: [@tryoo0607](https://github.com/tryoo0607)
- Docker Hub: [@tryoo0607](https://hub.docker.com/u/tryoo0607)