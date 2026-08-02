// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { act, cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { LanguageProvider } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import Navbar from "./Navbar";

vi.mock("../lib/api", () => ({ fetchJSON: vi.fn() }));

const feedItems = [
  { id: 42, productType: "block", stoneName: "Abbasabad Travertine", quantity: 24, unit: "ton" },
  { id: 43, productType: "finished", stoneName: "Black Marble", quantity: 12, unit: "ton" }
];

const renderNavbar = () => render(
  <LanguageProvider initialLang="en">
    <MemoryRouter>
      <Navbar />
    </MemoryRouter>
  </LanguageProvider>
);

describe("navbar live offers", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    fetchJSON.mockImplementation((path) => {
      if (path.startsWith("/api/ads/live-feed")) {
        return Promise.resolve({ data: { items: feedItems } });
      }
      return Promise.resolve({ data: {} });
    });
    window.matchMedia = vi.fn().mockImplementation(() => ({
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn()
    }));
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
    vi.useRealTimers();
  });

  it("rotates every six seconds and links to the real listing", async () => {
    renderNavbar();
    await act(async () => Promise.resolve());

    const firstLink = screen.getByRole("link", { name: /24 tons.*Abbasabad Travertine Block/i });
    expect(firstLink).toHaveAttribute("href", "/ads/42");

    await act(async () => vi.advanceTimersByTime(6250));
    expect(screen.getByRole("link", { name: /12 tons.*Black Marble Finished stone/i })).toHaveAttribute("href", "/ads/43");
  });

  it("pauses rotation while hovered or focused", async () => {
    renderNavbar();
    await act(async () => Promise.resolve());

    const link = screen.getByRole("link", { name: /Abbasabad Travertine Block/i });
    fireEvent.mouseEnter(link);
    await act(async () => vi.advanceTimersByTime(12000));
    expect(screen.getByRole("link", { name: /Abbasabad Travertine Block/i })).toBeInTheDocument();

    fireEvent.mouseLeave(link);
    await act(async () => vi.advanceTimersByTime(6250));
    const secondLink = screen.getByRole("link", { name: /Black Marble Finished stone/i });
    fireEvent.focus(secondLink);
    await act(async () => vi.advanceTimersByTime(12000));
    expect(screen.getByRole("link", { name: /Black Marble Finished stone/i })).toBeInTheDocument();
  });

  it("does not auto-rotate when reduced motion is requested", async () => {
    window.matchMedia = vi.fn().mockImplementation(() => ({
      matches: true,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn()
    }));
    renderNavbar();
    await act(async () => Promise.resolve());

    await act(async () => vi.advanceTimersByTime(20000));
    expect(screen.getByRole("link", { name: /Abbasabad Travertine Block/i })).toBeInTheDocument();
  });

  it("links to the Trade Board without fabricated data when the feed is empty", async () => {
    fetchJSON.mockResolvedValue({ data: { items: [] } });
    renderNavbar();
    await act(async () => Promise.resolve());

    expect(screen.getByRole("link", { name: "View the latest stone market listings" })).toHaveAttribute("href", "/ads");
  });

  it("keeps a neutral fallback when the feed request fails", async () => {
    fetchJSON.mockRejectedValue(new Error("offline"));
    renderNavbar();
    await act(async () => Promise.resolve());

    expect(screen.getByRole("link", { name: "Latest from the stone market" })).toHaveAttribute("href", "/ads");
  });
});
