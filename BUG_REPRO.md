# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	pharmacycounter/cmd/pharmacy	[no test files]
?   	pharmacycounter/config	[no test files]
--- FAIL: TestPharmacyNumberClaimedOnce (0.00s)
    pharmacy_test.go:153: number P001 claimed 2 times
FAIL
FAIL	pharmacycounter	0.071s
ok  	pharmacycounter/audit	0.003s
ok  	pharmacycounter/model	0.005s
ok  	pharmacycounter/persistence	0.024s
ok  	pharmacycounter/queue	0.004s
ok  	pharmacycounter/report	0.004s
ok  	pharmacycounter/service	0.023s
?   	pharmacycounter/web	[no test files]
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/pharmacy): exit `0`
- Frontend build (web): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/pharmacy): exit `0`
- Frontend build (web): exit `0`
