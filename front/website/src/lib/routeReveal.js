export const ROUTE_REVEAL = {
  tileCount: 3,
  tileStartDelay: 0.1,
  tileDuration: 0.55,
  tileDurationReduced: 0.30,
  mobileDurationMultiplier: 0.72,
  mobileViewportMax: 900,
  tileEase: "sine.inOut"
};

const shouldUseMobileTiming = () => {
  if (typeof window === "undefined") return false;
  const coarsePointer = window.matchMedia?.("(pointer: coarse)")?.matches;
  if (coarsePointer) return true;
  return window.innerWidth <= ROUTE_REVEAL.mobileViewportMax;
};

const getDurationMultiplier = () =>
  shouldUseMobileTiming() ? ROUTE_REVEAL.mobileDurationMultiplier : 1;

export const getTileStartDelay = () =>
  ROUTE_REVEAL.tileStartDelay * getDurationMultiplier();

export const getTileDuration = (reducedMotion = false) =>
  (reducedMotion ? ROUTE_REVEAL.tileDurationReduced : ROUTE_REVEAL.tileDuration) * getDurationMultiplier();

export const getTileCompletionTime = (index, reducedMotion = false) =>
  getTileDuration(reducedMotion) * (index + 1);
