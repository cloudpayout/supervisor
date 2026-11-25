<div align="center">
  <img width="400" height="400" alt="logo-with-supervisor-subtitle -cropped-modified" src="https://github.com/user-attachments/assets/ee03fb1e-2f78-4cae-a4c6-91dd45972a72" />
  <h1>👁️ Cloudpayout Supervisor</h1>
  <h3><em>Official Cloudpayout Supervisor to initialize automations on popular Casino games (Aviator, Spaceman, Dino)</em></h3>
</div>

<p align="center">
  <a href="https://github.com/cloudpayout/supervisor/releases"><img src="https://img.shields.io/github/v/release/cloudpayout/supervisor" alt="GitHub release"></a>
</p>

Cloudpayout Supervisor is the official CLI program to run automations on popular casino games such as Aviator, Spaceman, Dino, and so on.
Supervisor monitors automations by running <a href="https://github.com/chromedp/chromedp">chromedp</a> instances and running the necessary code to run these automations.

An `onewin` session start opens Aviator, chooses the game-balance currency, and activates play or demo. The instance stays up once the game is ready.

## Development (spec 017 headless shell)

Default headless runs use **chrome-headless-shell** from the repo, not system Chrome. Before `supervisor start --headless` or headless integration tests:

```bash
./scripts/fetch-chrome-headless-shell.sh
```

See [specs/017-optimizing-browsers-4/quickstart.md](specs/017-optimizing-browsers-4/quickstart.md). Headed `start` still uses `CHROME_PATH` or Chrome on `PATH`.
