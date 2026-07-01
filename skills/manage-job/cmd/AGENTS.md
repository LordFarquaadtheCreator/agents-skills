# Agent Instructions — cmd package

## Purpose

Cobra command definitions. Each file exports a `*cobra.Command` var. Commands are thin wrappers — they parse flags/args, build maps, call `appscript` methods, print results.

## Files

- `get.go` — `GetCmd`: fetches job applications, accepts key-value pairs as args for filtering
- `track.go` — `TrackCmd`: creates new job application, 4 required positional args + optional email/phone/notes
- `patch.go` — `PatchCmd`: updates existing row, requires `--matchBy` and `--update` JSON flags
- `delete.go` — `DeleteCmd`: deletes row, requires `--matchBy` JSON flag

## Flag registration

`--matchBy` and `--update` flags are registered in `main.go`, not in command files:

```go
cmd.PatchCmd.Flags().String("matchBy", "", "...")
cmd.PatchCmd.Flags().String("update", "", "...")
cmd.DeleteCmd.Flags().String("matchBy", "", "...")
```

## Adding a new command

1. Create `<name>.go` in this package
2. Export `<Name>Cmd *cobra.Command`
3. Import `manage-job/appscript`, call `appscript.NewAppScript()`
4. Register in `main.go`: `rootCmd.AddCommand(cmd.<Name>Cmd)`
5. Add any flags in `main.go`
