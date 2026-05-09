import { useEffect, useMemo, useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import { gsap } from "gsap";
import { ROUTE_REVEAL, getTileDuration } from "../lib/routeReveal";

const prefersReducedMotion = () => {
  if (typeof window === "undefined" || !window.matchMedia) return false;
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
};

export default function RouteRevealOverlay() {
  const location = useLocation();
  const [visible, setVisible] = useState(false);
  const panelRefs = useRef([]);
  const timelineRef = useRef(null);

  const routeKey = useMemo(() => `${location.pathname}${location.search}${location.hash}`, [location.pathname, location.search, location.hash]);

  useEffect(() => {
    setVisible(true);
    const raf = window.requestAnimationFrame(() => {
      const panels = panelRefs.current.filter(Boolean);
      if (!panels.length) {
        setVisible(false);
        return;
      }
      const reduceMotion = prefersReducedMotion();
      const tileDuration = getTileDuration(reduceMotion);

      gsap.set(panels, { yPercent: 0, force3D: true });
      timelineRef.current = gsap.timeline({ onComplete: () => setVisible(false) });
      timelineRef.current.to({}, { duration: reduceMotion ? 0.06 : ROUTE_REVEAL.tileStartDelay });
      panels.forEach((panel) => {
        timelineRef.current.to(panel, {
          yPercent: -110,
          duration: tileDuration,
          ease: ROUTE_REVEAL.tileEase,
          force3D: true
        });
      });
    });

    return () => {
      window.cancelAnimationFrame(raf);
      timelineRef.current?.kill();
    };
  }, [routeKey]);

  if (!visible) return null;

  return (
    <div className="route-overlay" aria-hidden>
      {Array.from({ length: ROUTE_REVEAL.tileCount }).map((_, index) => (
        <div
          key={`route-tile-${index}`}
          ref={(element) => {
            panelRefs.current[index] = element;
          }}
          className={`route-overlay__panel tile-${index + 1}`}
        />
      ))}
    </div>
  );
}
