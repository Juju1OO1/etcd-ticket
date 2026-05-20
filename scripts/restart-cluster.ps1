#!/usr/bin/env pwsh
# Restart any stopped etcd containers so the cluster is back to 3 nodes

$containers = @("etcd1", "etcd2", "etcd3")

Write-Host "=== Checking and starting stopped nodes ===" -ForegroundColor Cyan
foreach ($c in $containers) {
    $running = docker inspect -f '{{.State.Running}}' $c 2>$null
    if ($running -ne "true") {
        Write-Host "  Starting $c..." -ForegroundColor Yellow
        docker start $c | Out-Null
        Write-Host "  [OK] $c started" -ForegroundColor Green
    } else {
        Write-Host "  $c already running (skip)" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "=== Waiting for nodes to rejoin (3s) ===" -ForegroundColor Cyan
Start-Sleep -Seconds 3

Write-Host ""
& "$PSScriptRoot\show-cluster.ps1"
