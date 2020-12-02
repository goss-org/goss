param(
    # Start port scanning at
    [int] $startAt = 1025,
    # End port scanning at
    [int] $endAt = 65535
)
for ($port=$startAt; $port -lt $endAt; $port++) {
    $listener = New-Object System.Net.Sockets.TcpListener([System.Net.IPAddress]::Any, $port)
    try {
        $listener.Start()
        write-output "$port"
        break
    }
    catch {
        write-host "$port busy"
    }
    finally {
        $listener.Stop()
    }
}
