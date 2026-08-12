// @vitest-environment jsdom

import { fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialog, Pagination, PersianDate } from "./OperationalUI";

describe("production-readiness UI primitives", () => {
  it("paginates without passing the last page", () => {
    const onChange = vi.fn();
    render(<Pagination page={2} pageSize={25} total={51} onChange={onChange} />);
    fireEvent.click(screen.getByRole("button", { name: "بعدی" }));
    expect(onChange).toHaveBeenCalledWith(3);
  });

  it("requires a reason before a destructive action", () => {
    const onConfirm = vi.fn();
    render(<ConfirmDialog open title="لغو" description="پیامد عملیات" requireReason reason="" onReason={() => {}} onCancel={() => {}} onConfirm={onConfirm} />);
    expect(screen.getByRole("button", { name: "تأیید" })).toBeDisabled();
  });

  it("renders a standard UTC timestamp with the Persian calendar", () => {
    render(<PersianDate value="2026-08-11T12:00:00Z" />);
    const time = screen.getByRole("time");
    expect(time).toHaveAttribute("dateTime", "2026-08-11T12:00:00.000Z");
    expect(time.textContent).not.toBe("");
  });
});
