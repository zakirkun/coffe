import "@fontsource-variable/jetbrains-mono"
import "@fontsource-variable/manrope"
import { StrictMode } from "react"
import { createRoot } from "react-dom/client"

import { App } from "./app"
import { PrimitiveShowcase } from "./components/showcase"
import "./styles/tokens.css"
import "./styles/base.css"
import "./styles/components.css"
import "./styles/forms.css"
import "./styles/app.css"
import "./styles/showcase.css"

if (import.meta.env.DEV) {
  await Promise.all([import("react-grab"), import("react-scan")])
}

const rootElement = document.getElementById("root")
if (rootElement === null) {
  throw new Error("Root element not found")
}

createRoot(rootElement).render(
  <StrictMode>
    {new URLSearchParams(window.location.search).has("showcase") ? <PrimitiveShowcase /> : <App />}
  </StrictMode>,
)
