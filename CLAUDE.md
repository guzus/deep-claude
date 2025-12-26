# Continuous Claude

## Debugging New Features

To test new features on a remote machine:

1. **Build the Linux binary:**
   ```bash
   make build-linux
   ```

2. **Copy to remote machine:**
   ```bash
   make scp-linux
   ```

3. **Test on remote:**
   ```bash
   ssh continuous-claude-vm "mkdir -p ~/test-folder && chmod +x ~/continuous-claude-linux && cd ~/test-folder && ~/continuous-claude-linux -p 'test' --max-runs 1 --dry-run"
   ```

Use `--dry-run` to simulate without making actual changes. The `--max-runs 1` flag limits execution to a single iteration for quick testing.
