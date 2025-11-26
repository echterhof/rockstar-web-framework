# Generate self-signed TLS certificates for QUIC server example
# These certificates are for development/testing purposes only

param(
    [string]$CertFile = "cert.pem",
    [string]$KeyFile = "key.pem"
)

Write-Host "🔐 Generating TLS certificates for QUIC server..." -ForegroundColor Cyan
Write-Host "   Certificate: $CertFile"
Write-Host "   Private Key: $KeyFile"
Write-Host ""

# Check if OpenSSL is available
$opensslPath = Get-Command openssl -ErrorAction SilentlyContinue

if (-not $opensslPath) {
    Write-Host "❌ OpenSSL not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install OpenSSL:" -ForegroundColor Yellow
    Write-Host "  • Download from: https://slproweb.com/products/Win32OpenSSL.html"
    Write-Host "  • Or install via Chocolatey: choco install openssl"
    Write-Host "  • Or install via Scoop: scoop install openssl"
    Write-Host ""
    exit 1
}

# Generate self-signed certificate valid for 365 days
& openssl req -x509 `
    -newkey rsa:4096 `
    -keyout $KeyFile `
    -out $CertFile `
    -days 365 `
    -nodes `
    -subj '/CN=localhost' `
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Certificates generated successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "⚠️  These are self-signed certificates for development only." -ForegroundColor Yellow
    Write-Host "   Do NOT use in production!"
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "❌ Failed to generate certificates!" -ForegroundColor Red
    Write-Host ""
    exit 1
}
