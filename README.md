# bkptool

Ergonomic local backups for day-to-day file editing.

`bkptool` is a CLI for creating, restoring, and inspecting backup history of files/directories on your filesystem.

## Why this exists

You edit config files, scripts, and dotfiles all day.

Classic workflow:

```bash
cp ~/.config/app-name/config.yaml ~/.config/app-name/config.yaml.bkp
vim ~/.config/app-name/config.yaml
# regret happens
mv ~/.config/app-name/config.yaml.bkp ~/.config/app-name/config.yaml
```

It works, but it is manual, error-prone, and history depth is usually one backup.

`bkptool` turns this into a first-class, repeatable workflow with history and diffs.

## Core goals

- Make backup/restore a single command away.
- Keep backups in a per-target **stack**.
- Default restore behavior should be **LIFO** (pop most recent backup).
- Allow restoring by index (not only latest).
- Provide history-aware diffs similar to git ergonomics.
- Support both files and directories.

## Planned CLI

### Backup

```bash
bkptool backup <path>
# likely alias: bkp <path>
```

Creates a new backup snapshot for the target path and pushes it onto that target's backup stack.

### Restore

```bash
bkptool restore <path>
# likely alias: restore <path>
```

Restores from backup history.

Default behavior:
- Restores latest backup for `<path>` (stack pop / LIFO).

Extended behavior:
- Restore by index (e.g. `--index N`) for non-top entries.
- Optional non-destructive restore mode (`--keep`) can be added later.

### Diff

```bash
bkptool diff <path>
```

Shows differences across backup history.

Expected modes (MVP + near-term):
- working copy vs latest backup
- backup index A vs backup index B
- latest backup vs previous backup

## Mental model

Each tracked target path has its own backup history stack:

- `backup` => push new snapshot
- `restore` (default) => pop latest snapshot and restore it
- `restore --index N` => restore chosen snapshot (behavior around pop/remove should be explicit)
- `diff` => inspect changes between working copy and/or snapshots

## Suggested MVP behavior

1. `backup <path>`
   - Validate path exists
   - Snapshot file/dir into a managed store
   - Append metadata entry (timestamp, source path, hash, size, type)

2. `list <path>` (recommended command for visibility)
   - Show backup stack for target path with indices and timestamps

3. `restore <path>`
   - Default: restore latest and remove that backup from top (pop)
   - `--index N`: restore selected index (define whether this removes N or not)

4. `diff <path>`
   - Default: current working copy vs latest backup
   - Add index selectors later (`--from`, `--to`)

## Storage design (initial)

- Keep snapshots in a dedicated data dir (e.g. `~/.local/share/bkptool/` on Linux).
- Store metadata in a small local DB or structured files.
- Snapshots should be immutable once created.
- Track per-source-path history independently.

## Safety expectations

- No silent overwrite without explicit intent.
- Clear confirmations for destructive actions.
- Path normalization to avoid duplicate identities (`~/x`, `/home/user/x`, symlinks).
- Good error messages when target disappeared, permissions changed, or history is empty.

## Non-goals (for now)

- Cloud sync / remote backups.
- Full version-control replacement.
- Collaborative workflows.

## Initial roadmap / tasks

### Phase 1 (MVP)
- [ ] Implement `backup <path>` for files
- [ ] Implement `restore <path>` (latest/LIFO)
- [ ] Implement `restore --index N`
- [ ] Implement `list <path>`
- [ ] Implement `diff <path>` (working vs latest)
- [ ] Add tests for stack semantics

### Phase 2
- [ ] Directory backup support with recursive snapshots
- [ ] `diff --from/--to` between backup indices
- [ ] Better output UX (colored diffs, concise status)
- [ ] Optional aliases (`bkp`, `restore`) in install docs

### Phase 3
- [ ] Retention policies / pruning
- [ ] Tags/notes for snapshots
- [ ] Optional compression and deduplication

## Example workflow

```bash
bkptool backup ~/.config/app-name/config.yaml
vim ~/.config/app-name/config.yaml
bkptool diff ~/.config/app-name/config.yaml
bkptool restore ~/.config/app-name/config.yaml
```

## Status

Early development. Interface and internals may change while core semantics are being nailed down.
