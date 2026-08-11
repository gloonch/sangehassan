// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import ChangePassword from "./ChangePassword";

const fetchJSON = vi.fn();
const refreshUser = vi.fn();

vi.mock("../lib/api", () => ({
  fetchJSON: (...args) => fetchJSON(...args)
}));

vi.mock("../lib/auth", () => ({
  useAuth: () => ({ refreshUser })
}));

describe("temporary password change", () => {
  beforeEach(() => {
    fetchJSON.mockReset();
    refreshUser.mockReset();
  });

  it("shows the real current-password error instead of a length error", async () => {
    const requestError = new Error("current password is incorrect");
    requestError.code = "CURRENT_PASSWORD_INVALID";
    fetchJSON.mockRejectedValue(requestError);

    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <ChangePassword />
      </MemoryRouter>
    );

    await user.type(screen.getByPlaceholderText("رمز موقت"), "Temporary1!");
    await user.type(screen.getByPlaceholderText("رمز جدید (حداقل ۸ نویسه)"), "SecurePass1!");
    await user.click(screen.getByRole("button", { name: "ذخیره رمز جدید" }));

    expect(await screen.findByText(/رمز موقت فعلی صحیح نیست/)).toBeInTheDocument();
    expect(refreshUser).not.toHaveBeenCalled();
  });
});
