import React from "react";
import { renderToString } from "react-dom/server";
import { StaticRouter } from "react-router-dom/server";
import App from "./App";
import { getLanguageFromPath, LanguageProvider } from "./lib/i18n";
import { PrerenderDataProvider } from "./lib/prerenderData";

export function render(url, prerenderData = null) {
  return renderToString(
    <React.StrictMode>
      <LanguageProvider initialLang={getLanguageFromPath(url)}>
        <PrerenderDataProvider data={prerenderData}>
          <StaticRouter location={url}>
            <App />
          </StaticRouter>
        </PrerenderDataProvider>
      </LanguageProvider>
    </React.StrictMode>
  );
}
