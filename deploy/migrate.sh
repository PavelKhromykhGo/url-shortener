set -euo pipefail

echo "[migrator] starting..."

for i in $(seq 1 60); do
  set +e
  out=$(migrate -path /migrations -database "$POSTGRES_DSN" up 2>&1)
  code=$?
  set -e

  if [ $code -eq 0 ]; then
    echo "[migrator] success"
    echo "$out"
    exit 0
  fi

  if echo "$out" | grep -qiE "connection refused|connect: cannot assign requested address|dial tcp|no such host|timeout|EOF"; then
    echo "[migrator] postgres not ready yet... ($i/60)"
    sleep 1
    continue
  fi

  echo "[migrator] failed:"
  echo "$out"
  exit 1
done

echo "[migrator] postgres did not become ready in time"
exit 1
