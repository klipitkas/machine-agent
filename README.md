# machine-agent — one-pager

Static landing page for [klipitkas/machine-agent](https://github.com/klipitkas/machine-agent).

## Deploy to GitHub Pages

1. Create a new public repo on GitHub (e.g. `machine-agent-site`).
2. Upload `index.html` (and this README) to the repo root.
   - Via web: repo → **Add file → Upload files** → drag `index.html` → Commit.
   - Via CLI:
     ```bash
     git init
     git add .
     git commit -m "initial"
     git branch -M main
     git remote add origin https://github.com/<you>/<repo>.git
     git push -u origin main
     ```
3. Repo → **Settings → Pages**.
4. **Source:** Deploy from a branch. **Branch:** `main` / `/ (root)`. **Save.**
5. Wait ~30s. Your site is live at `https://<you>.github.io/<repo>/`.

## Custom domain (optional)

Settings → Pages → Custom domain → enter your domain → add a `CNAME` DNS record pointing at `<you>.github.io`.

## Local preview

Just open `index.html` in a browser, or:
```bash
python3 -m http.server 8080
```
