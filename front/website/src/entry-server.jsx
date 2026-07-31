import React from "react";
import { PassThrough } from "node:stream";
import { renderToPipeableStream } from "react-dom/server";
import { StaticRouter } from "react-router-dom/server";
import App from "./App";
import { getLanguageFromPath, LanguageProvider } from "./lib/i18n";
import { PrerenderDataProvider } from "./lib/prerenderData";

export function render(url, prerenderData = null) {
  return new Promise((resolve, reject) => {
    let didError = false;
    const output = new PassThrough();
    let html = "";
    output.setEncoding("utf8");
    output.on("data", (chunk) => { html += chunk; });
    output.on("end", () => didError ? reject(new Error(`SSR failed for ${url}`)) : resolve(html));
    output.on("error", reject);

    const stream = renderToPipeableStream(
      <React.StrictMode>
        <LanguageProvider initialLang={getLanguageFromPath(url)}>
          <PrerenderDataProvider data={prerenderData}>
            <StaticRouter location={url}>
              <App />
            </StaticRouter>
          </PrerenderDataProvider>
        </LanguageProvider>
      </React.StrictMode>,
      {
        onAllReady() {
          stream.pipe(output);
        },
        onError(error) {
          didError = true;
          reject(error);
        }
      }
    );
  });
}
