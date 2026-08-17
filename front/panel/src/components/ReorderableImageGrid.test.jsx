// @vitest-environment jsdom

import { fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, expect, it, vi } from "vitest";
import ReorderableImageGrid from "./ReorderableImageGrid";

describe("ReorderableImageGrid", () => {
  it("shows slideshow order and allows configured media to be selected", () => {
    const onSelect = vi.fn();
    render(
      <ReorderableImageGrid
        images={["/images/one.webp", "/images/two.webp"]}
        selectedIndex={0}
        onSelect={onSelect}
        showOrderNumbers
        labels={{ slide: "اسلاید" }}
      />
    );

    expect(screen.getByText("اسلاید 1 / 2")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("img", { name: "اسلاید 2" }));
    expect(onSelect).toHaveBeenCalledWith(1);
  });

  it("does not show destructive controls for read-only default media", () => {
    render(
      <ReorderableImageGrid
        images={["/images/default.webp"]}
        showPreview={false}
        showOrderNumbers
        labels={{ slide: "اسلاید", remove: "حذف" }}
      />
    );

    expect(screen.queryByRole("button", { name: "حذف" })).not.toBeInTheDocument();
  });
});
