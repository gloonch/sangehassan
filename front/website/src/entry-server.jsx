import React from "react";
import { renderToString } from "react-dom/server";
import { StaticRouter } from "react-router-dom/server";
import App from "./App";
import { LanguageProvider } from "./lib/i18n";

export function render(url) {
  return renderToString(
    <React.StrictMode>
      <LanguageProvider>
        <StaticRouter location={url}>
          <App />
        </StaticRouter>
      </LanguageProvider>
    </React.StrictMode>
  );
}
