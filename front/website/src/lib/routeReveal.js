export const ROUTE_REVEAL = {
  tileCount: 3,
  tileStartDelay: 0.2,
  tileDuration: 1.55,
  tileDurationReduced: 0.55,
  tileEase: "sine.inOut"
};

export const getTileDuration = (reducedMotion = false) =>
  reducedMotion ? ROUTE_REVEAL.tileDurationReduced : ROUTE_REVEAL.tileDuration;

export const getTileCompletionTime = (index, reducedMotion = false) =>
  getTileDuration(reducedMotion) * (index + 1);
