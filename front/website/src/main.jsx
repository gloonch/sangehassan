import React from "react";
import { createRoot, hydrateRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App";
import { LanguageProvider } from "./lib/i18n";
import { getBrowserPrerenderData, PrerenderDataProvider } from "./lib/prerenderData";
import "./styles.css";

const prerenderData = getBrowserPrerenderData();

const app = (
  <React.StrictMode>
    <LanguageProvider>
      <PrerenderDataProvider data={prerenderData}>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </PrerenderDataProvider>
    </LanguageProvider>
  </React.StrictMode>
);

const root = document.getElementById("root");

if (root.hasChildNodes()) {
  hydrateRoot(root, app);
} else {
  createRoot(root).render(app);
}
