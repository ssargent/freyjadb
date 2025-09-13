FreyjaDB CLI Evolution: From init/serve to up/service

1. Purpose of this document

This document captures:

Current behavior of FreyjaDB’s CLI setup and start sequence.

Desired behavior for a simplified, modern UX (freyja up, freyja service …).

Implementation notes and guidance to help another LLM (or developer) implement the new commands.

2. Current State

Today a user must initialize a system store and then serve it manually.
Based on the current System Store User Guide:

Initialization
freyja init \
  --system-key <SYSTEM_KEY> \
  --system-api-key <SYSTEM_API_KEY> \
  --data-dir <DATA_DIR>

Creates the system store in the specified --data-dir.

Requires the caller to provide:

--system-key (used to protect the system store)

--system-api-key (admin-level API key)

Optional flags configure encryption and other parameters.

Starting the server
freyja serve \
  --api-key <CLIENT_API_KEY> \
  --system-key <SYSTEM_KEY> \
  --data-dir <DATA_DIR> \
  [--enable-encryption] \
  [--port <PORT>]

Starts the REST API server.

Requires re-supplying --system-key and --api-key each time.

Many secrets are passed on the command line, exposing them to shell history.

There is no built-in way to install FreyjaDB as a long-running daemon.

Pain points

Repetition: Users retype the same keys and flags.

Security: Secrets on the command line risk exposure.

Complexity: No single “one-liner” to get from zero to a running server.

No built-in service management: systemd setup is manual.

3. Target State
3.1 Key Goals

Bootstrap once, run anywhere: a single command to initialize and launch.

Config-driven: persist all settings and secrets in a secure config file.

Service management: native helpers to install/manage a systemd service.

4. Proposed CLI Design
4.1 freyja up

One-shot bootstrap and start.

Purpose: Detect first run, create config + keys if missing, then start server.

freyja up [options]

Options:

Flag Description Default
--data-dir Path to data store ./data
--port Port to listen on 8080
--bind Address to bind 127.0.0.1
--config Path to config file OS-specific default
--non-interactive Skip prompts, use defaults false
--print-keys Print generated API keys false

Behavior:

If config file is absent:

Create config directory.

Generate system_key, system_api_key, and a default client_api_key.

Save them securely (0600 permissions) in config.yaml or keys.json.

If config exists: validate and reuse keys.

Start the server in the foreground using the config.

Idempotent: Safe to run repeatedly.

4.2 freyja service

Manage a systemd service.

Subcommands:

Command Effect
freyja service install Install systemd unit and enable it
freyja service start Start the service immediately
freyja service stop Stop the service
freyja service restart Restart the service
freyja service status Show status via systemctl status
freyja service logs Tail logs via journalctl -u freyja
freyja service uninstall Remove systemd unit

install options:

freyja service install \
  [--data-dir /var/lib/freyjadb] \
  [--config /etc/freyja/config.yaml] \
  [--user freyja] \
  [--port 8080]

Behavior:

Ensure config and keys exist (run same bootstrap as freyja up --non-interactive).

Create a systemd unit file (template below).

Enable and (optionally) start the service:

systemctl enable --now freyja.service

Print next steps.

4.3 Systemd unit template
[Unit]
Description=FreyjaDB Server
After=network-online.target
Wants=network-online.target

[Service]
User=freyja
Group=freyja
ExecStart=/usr/local/bin/freyja serve --config /etc/freyja/config.yaml
Restart=on-failure
NoNewPrivileges=true
UMask=0077
ReadWritePaths=/var/lib/freyjadb /etc/freyja

[Install]
WantedBy=multi-user.target

5. Implementation Guidance for LLM

When implementing:

Refactor CLI

Use a subcommand pattern (up, service, etc.).

Keep existing init and serve for backward compatibility.

Bootstrap library

Factor out logic that:

Checks for existing config.

Generates keys securely.

Writes config with correct permissions.

Shared by both up and service install.

Config file

YAML or TOML:

data_dir: /var/lib/freyjadb
port: 8080
bind: 127.0.0.1
security:
  system_key: auto
  system_api_key: auto
  client_api_key: auto
logging:
  level: info

Keys may be stored in a separate keys.json (0600) and referenced from the YAML.

Systemd helper

For service install:

Detect distro path for systemd units (/etc/systemd/system).

Write the unit file (template above) with correct user/group.

Run systemctl daemon-reload and enable.

Provide wrappers for start/stop/restart/status/logs.

Backward compatibility

Keep freyja init and freyja serve unchanged for now.

Mark them as “advanced/manual mode” in docs.

Testing

Unit test config generation and key rotation.

Integration test: freyja up → curl test endpoint.

Systemd integration test using a container with systemd enabled.

Docs

Update README with a “Zero-to-PUT in 60s” quickstart:

brew install freyja   # or package install
freyja up
curl -H "X-API-Key: $(freyja print client_api_key)" \
  <http://127.0.0.1:8080/api/v1/kv/hello> -d "world"

6. Migration Plan
Step Action
1 Implement config package and secure key generation
2 Create up command that uses the config package
3 Build service subcommands and systemd unit template
4 Mark init/serve as legacy but keep functional
5 Update documentation and release notes
7. Future considerations

Optional TUI wizard inside freyja up for interactive first-run.

Key rotation command: freyja admin rotate --what client_api_key.

Docker image that automatically runs freyja up on container start.

End of Document

This single markdown file can be given to another LLM or developer as the specification for implementing the next-generation CLI for FreyjaDB.
