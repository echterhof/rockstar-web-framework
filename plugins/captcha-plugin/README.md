# CAPTCHA Plugin

A compile-time plugin for the Rockstar Web Framework that provides CAPTCHA generation and validation for bot protection.

## Features

- **CAPTCHA Generation**: Create random text-based CAPTCHA challenges
- **Validation**: Validate user responses against stored challenges
- **Expiration**: Automatic expiration of CAPTCHA challenges
- **Attempt Limiting**: Limit the number of validation attempts
- **Path Protection**: Protect specific paths with CAPTCHA validation
- **Service Export**: Export CaptchaService for use by other plugins
- **Configurable**: Highly configurable challenge length, expiry, and characters

## Configuration

```yaml
plugins:
  captcha-plugin:
    enabled: true
    config:
      enabled: true
      captcha_length: 6
      captcha_expiry: "5m"
      protected_paths:
        - "/api/login"
        - "/api/register"
        - "/api/contact"
      protected_methods:
        - "POST"
      case_sensitive: false
      allowed_characters: "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
      max_attempts: 3
    permissions:
      router: true
```

## Usage

### Automatic Protection

The plugin automatically protects configured paths. Clients must include CAPTCHA headers:

```http
POST /api/login HTTP/1.1
X-Captcha-ID: <captcha_id>
X-Captcha-Answer: <user_answer>
```

### Using the Exported Service

Other plugins can import and use the CaptchaService:

```go
// Import the service
service, err := ctx.ImportService("captcha-plugin", "CaptchaService")
if err != nil {
    return err
}
captchaService := service.(*CaptchaService)

// Generate a CAPTCHA
id, challenge, err := captchaService.GenerateCaptcha()
if err != nil {
    return err
}

// Display challenge to user, then validate their answer
valid, err := captchaService.ValidateCaptcha(id, userAnswer)
if err != nil {
    return err
}
```

## API Endpoints

You can create endpoints to generate CAPTCHAs:

```go
// GET /api/captcha - Generate a new CAPTCHA
router.GET("/api/captcha", func(ctx pkg.Context) error {
    service, _ := pluginCtx.ImportService("captcha-plugin", "CaptchaService")
    captchaService := service.(*CaptchaService)
    
    id, challenge, err := captchaService.GenerateCaptcha()
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{"error": err.Error()})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "id": id,
        "challenge": challenge,
    })
})
```

## How It Works

1. **Generation**: Plugin generates a random string using configured characters
2. **Storage**: Challenge is stored with an ID, expiry time, and attempt counter
3. **Validation**: When a request arrives, the plugin checks the CAPTCHA headers
4. **Verification**: Answer is compared with stored challenge (case-insensitive by default)
5. **Cleanup**: Expired CAPTCHAs are automatically cleaned up in the background

## Security Features

- **Expiration**: CAPTCHAs expire after configured duration (default 5 minutes)
- **Attempt Limiting**: Maximum attempts per CAPTCHA (default 3)
- **One-Time Use**: Valid CAPTCHAs are deleted after successful validation
- **Ambiguous Characters Excluded**: Default character set excludes 0, O, 1, I, etc.
- **Random Generation**: Cryptographically secure random generation

## Events

The plugin publishes the following events:

- `captcha.generated` - When a new CAPTCHA is generated
- `captcha.validated` - When a CAPTCHA is successfully validated
- `captcha.failed` - When CAPTCHA validation fails

## Best Practices

1. **Use HTTPS**: Always use HTTPS to protect CAPTCHA challenges in transit
2. **Rate Limiting**: Combine with rate limiting for additional protection
3. **Adjust Difficulty**: Configure length and expiry based on your needs
4. **Monitor Events**: Subscribe to CAPTCHA events for security monitoring
5. **Client Integration**: Provide clear UI for CAPTCHA challenges

## Example Client Integration

```javascript
// Fetch a CAPTCHA
const response = await fetch('/api/captcha');
const { id, challenge } = await response.json();

// Display challenge to user
displayCaptcha(challenge);

// Submit form with CAPTCHA
const formData = { /* form data */ };
await fetch('/api/login', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'X-Captcha-ID': id,
        'X-Captcha-Answer': userAnswer
    },
    body: JSON.stringify(formData)
});
```

## License

This plugin is part of the Rockstar Web Framework and is provided under the same license.
