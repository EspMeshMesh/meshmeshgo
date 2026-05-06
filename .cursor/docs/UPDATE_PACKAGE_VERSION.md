# Procedure to update and publish new MeshMeshGo revision

## Procedure

1. Wanted revision always has the format v1.2.3 with a leading v
2. Change `programRevision` constant in main.go to match the wanted revision. Use the numeric version only (no leading `v`), e.g. `1.2.3`.
2. Commit the changes with message `bump to 1.2.3` where `1.2.3` is the new version.
3. Add an annotated Git tag `v1.2.3` with the same digits. Use the tag name as the tag message (e.g. `-m "v1.2.3"`).
4. Push the main branch and tags to `origin`.

## Useful commands

```bash
git tag -a v1.2.3 -m "v1.2.3"
git push origin main
git push origin v1.2.3
```

Replace `1.2.3` / `v1.2.3` and `main` with the version and branch you are releasing.