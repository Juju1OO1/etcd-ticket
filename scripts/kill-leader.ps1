#!/usr/bin/env pwsh
# Find the current Raft Leader and kill it, then verify Raft re-election

$endpoints = "http://etcd1:2379,http://etcd2:2379,http://etcd3:2379"
$containers = @("etcd1", "etcd2", "etcd3")

# Find an alive container
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
    exit 1
}

# Find Leader via JSON output
# Each endpoint reports the leader_id it sees.
# The endpoint whose own member_id == leader_id is the actual Leader.
Write-Host "=== Step 1: Locate current Leader ===" -ForegroundColor Cyan
$json = docker exec $alive etcdctl --endpoints=$endpoints endpoint status -w json | ConvertFrom-Json

$leaderContainer = $null
foreach ($node in $json) {
    $memberId = "$($node.Status.header.member_id)"
    $leaderId = "$($node.Status.leader)"
    if ($memberId -eq $leaderId) {
        $leaderContainer = switch -Wildcard ($node.Endpoint) {
            "*etcd1*" { "etcd1" }
            "*etcd2*" { "etcd2" }
            "*etcd3*" { "etcd3" }
        }
        break
    }
}

if (-not $leaderContainer) {
    Write-Host "[ERROR] Cannot find Leader (election may be in progress). Retry in a few seconds." -ForegroundColor Red
    exit 1
}

Write-Host "  Current Leader: $leaderContainer" -ForegroundColor Yellow
Write-Host ""

# Kill the leader container
Write-Host "=== Step 2: Killing $leaderContainer ===" -ForegroundColor Red
docker kill $leaderContainer | Out-Null
Write-Host "  [OK] $leaderContainer stopped" -ForegroundColor Gray
Write-Host ""

# Wait for Raft re-election (default election timeout ~1s, give it 4s buffer)
Write-Host "=== Step 3: Waiting for Raft re-election (4s) ===" -ForegroundColor Cyan
for ($i = 4; $i -gt 0; $i--) {
    Write-Host "  $i..." -ForegroundColor Gray
    Start-Sleep -Seconds 1
}
Write-Host ""

# Show new cluster state
Write-Host "=== Step 4: New Cluster Status ===" -ForegroundColor Green
& "$PSScriptRoot\show-cluster.ps1"

Write-Host ""
Write-Host "=== Demo Notes ===" -ForegroundColor Magenta
Write-Host "  - $leaderContainer is down, cluster runs on 2 nodes" -ForegroundColor White
Write-Host "  - Raft elected a new Leader (quorum 2/3 satisfied)" -ForegroundColor White
Write-Host "  - Ticket tests should still PASS - cluster remains usable" -ForegroundColor White
Write-Host ""
Write-Host "  To restore the cluster: .\scripts\restart-cluster.ps1" -ForegroundColor Gray
