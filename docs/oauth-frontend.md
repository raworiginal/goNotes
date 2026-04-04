# Frontend OAuth Callback — Tutorial

After GitHub/Google redirects back to your Vite app, the URL looks like:

```
http://localhost:5173/#code=abc123xyz
```

Your React app needs to read that code, trade it for real tokens, and store them. Here's how.

---

## Step 1 — Create the callback component

Create a new file: `src/OAuthCallback.jsx`

This component runs once when it mounts, reads the URL fragment, and calls the exchange endpoint.

```jsx
import { useEffect, useState } from "react";

export default function OAuthCallback() {
  const [status, setStatus] = useState("Signing you in...");

  useEffect(() => {
    // window.location.hash is "#code=abc123" — slice off the "#" first
    const params = new URLSearchParams(window.location.hash.slice(1));
    const code = params.get("code");
    const error = params.get("error");

    if (error) {
      setStatus(`Login failed: ${error}`);
      return;
    }

    if (!code) {
      setStatus("No code found in URL.");
      return;
    }

    fetch("http://localhost:3000/auth/exchange", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code }),
    })
      .then((res) => {
        if (!res.ok) throw new Error("exchange failed");
        return res.json();
      })
      .then((data) => {
        // Store tokens — localStorage is simple and standard for learning projects.
        // In production you'd want the refresh token in an HttpOnly cookie instead.
        localStorage.setItem("access_token", data.access_token);
        localStorage.setItem("refresh_token", data.refresh_token);
        window.location.href = "/";
      })
      .catch(() => setStatus("Login failed. Please try again."));
  }, []);

  return <p>{status}</p>;
}
```

---

## Step 2 — Render the callback component when needed

Open `src/App.jsx`. At the top of your `App` function, check if the current URL has a `code` or `error` in the hash fragment. If it does, render `OAuthCallback` instead of the normal app.

```jsx
import OAuthCallback from "./OAuthCallback";

function App() {
  const params = new URLSearchParams(window.location.hash.slice(1));
  if (params.has("code") || params.has("error")) {
    return <OAuthCallback />;
  }

  // Your normal app goes below
  return (
    <div>
      <h1>goNotes</h1>
      <a href="http://localhost:3000/auth/github">Sign in with GitHub</a>
    </div>
  );
}

export default App;
```

---

## Step 3 — Test it

1. Make sure both servers are running:
   - Go backend: `make run` (port 3000)
   - Vite frontend: `npm run dev` (port 5173)

2. Open `http://localhost:5173` and click "Sign in with GitHub"

3. Authorize on GitHub — you'll be redirected back to `localhost:5173`

4. The page should briefly show "Signing you in..." then redirect to the home page

5. Open browser DevTools → Application → Local Storage → `http://localhost:5173`
   — you should see `access_token` and `refresh_token`

6. Test the token works:
   ```bash
   curl http://localhost:3000/notes \
     -H "Authorization: Bearer <paste access_token here>"
   ```
   You should get back an empty notes array `[]`, not a 401.

---

## How it all fits together

```
Browser visits /auth/github
      ↓
Go server redirects → GitHub
      ↓
User authorizes on GitHub
      ↓
GitHub redirects → /auth/github/callback (Go server)
      ↓
Go server creates one-time code, redirects → localhost:5173/#code=xyz
      ↓
OAuthCallback reads code from URL fragment
      ↓
POST /auth/exchange → Go server validates code, returns JWT + refresh token
      ↓
Tokens stored in localStorage, redirect to home
```
