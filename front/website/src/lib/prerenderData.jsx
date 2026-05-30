import { createContext, useContext } from "react";

const PrerenderDataContext = createContext(null);

export function PrerenderDataProvider({ data, children }) {
  return <PrerenderDataContext.Provider value={data || null}>{children}</PrerenderDataContext.Provider>;
}

export function getBrowserPrerenderData() {
  if (typeof window === "undefined") return null;
  return window.__SH_PRERENDER_DATA__ || null;
}

export function usePrerenderData(key) {
  const data = useContext(PrerenderDataContext);
  return key ? data?.[key] || null : data;
}
