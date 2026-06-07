$container = 'etcd1'
$ticketKey = 'area1/limit/current_limit'
Write-Host '正在檢查 etcd 剩餘票數...' -ForegroundColor Cyan
$stock = docker exec $container etcdctl get $ticketKey --print-value-only
if (-not $stock) { Write-Host '[ERROR] 找不到 Key:' $ticketKey -ForegroundColor Red; exit 1 }
Write-Host '目前剩餘票數:' $stock -ForegroundColor Yellow
if ([int]$stock -eq 0) { Write-Host '測試通過：票數精準歸零，無超賣！' -ForegroundColor Green } elseif ([int]$stock -lt 0) { Write-Host '測試失敗：發生超賣！' -ForegroundColor Red } else { Write-Host '票未售完，剩餘' $stock '張。' -ForegroundColor Magenta }
