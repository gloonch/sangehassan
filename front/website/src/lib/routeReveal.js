export const ROUTE_REVEAL = {
  tileCount: 3,
  tileStartDelay: 0.1,
  tileDuration: 0.55,
  tileDurationReduced: 0.30,
  tileEase: "sine.inOut"
};

export const getTileDuration = (reducedMotion = false) =>
  reducedMotion ? ROUTE_REVEAL.tileDurationReduced : ROUTE_REVEAL.tileDuration;

export const getTileCompletionTime = (index, reducedMotion = false) =>
  getTileDuration(reducedMotion) * (index + 1);
