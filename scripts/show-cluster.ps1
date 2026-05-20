#!/usr/bin/env pwsh
# Show etcd cluster status and identify the current Raft Leader

$endpoints = "http://etcd1:2379,http://etcd2:2379,http://etcd3:2379"
$containers = @("etcd1", "etcd2", "etcd3")

# Find an alive container to exec etcdctl through
$alive = $null
foreach ($c in $containers) {
    $running = docker inspect -f '{{.State.Running}}' $c 2>$null
    if ($running -eq "true") {
        $alive = $c
        break
    }
}

if (-not $alive) {
    Write-Host "[ERROR] No etcd node is running." -ForegroundColor Red
    Write-Host "Run: docker-compose -f backend\internal\etcd\etcd-cluster\docker-compose.yaml up -d" -ForegroundColor Yellow
    exit 1
}

Write-Host "=== Cluster Status (via $alive) ===" -ForegroundColor Cyan
docker exec $alive etcdctl --endpoints=$endpoints endpoint status -w table

Write-Host ""
Write-Host "=== Container Status ===" -ForegroundColor Cyan
foreach ($c in $containers) {
    $running = docker inspect -f '{{.State.Running}}' $c 2>$null
    if ($running -eq "true") {
        Write-Host "  $c : Running" -ForegroundColor Green
    } else {
        Write-Host "  $c : Stopped" -ForegroundColor Red
    }
}
