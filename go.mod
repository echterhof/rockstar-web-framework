module github.com/echterhof/rockstar-web-framework

go 1.25.4

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/echterhof/rockstar-web-framework/plugins/auth-plugin v0.0.0-00010101000000-000000000000
	github.com/echterhof/rockstar-web-framework/plugins/cache-plugin v0.0.0-00010101000000-000000000000
	github.com/echterhof/rockstar-web-framework/plugins/captcha-plugin v0.0.0-00010101000000-000000000000
	github.com/echterhof/rockstar-web-framework/plugins/logging-plugin v0.0.0-00010101000000-000000000000
	github.com/echterhof/rockstar-web-framework/plugins/storage-plugin v0.0.0-00010101000000-000000000000
	github.com/echterhof/rockstar-web-framework/plugins/template v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.9.3
	github.com/gorilla/websocket v1.5.3
	github.com/leanovate/gopter v0.2.11
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/microsoft/go-mssqldb v1.9.4
	github.com/quic-go/quic-go v0.57.0
	golang.org/x/crypto v0.45.0
	golang.org/x/net v0.47.0
	golang.org/x/sys v0.38.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)

replace github.com/echterhof/rockstar-web-framework => ./

// Plugin replacements (automatically managed by 'make discover-plugins')
// Add plugin replacements below this line
replace github.com/echterhof/rockstar-web-framework/plugins/auth-plugin => ./plugins/auth-plugin

replace github.com/echterhof/rockstar-web-framework/plugins/cache-plugin => ./plugins/cache-plugin

replace github.com/echterhof/rockstar-web-framework/plugins/captcha-plugin => ./plugins/captcha-plugin

replace github.com/echterhof/rockstar-web-framework/plugins/logging-plugin => ./plugins/logging-plugin

replace github.com/echterhof/rockstar-web-framework/plugins/storage-plugin => ./plugins/storage-plugin

replace github.com/echterhof/rockstar-web-framework/plugins/template => ./plugins/template
