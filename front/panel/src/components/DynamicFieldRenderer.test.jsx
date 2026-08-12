// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import DynamicFieldRenderer from "./DynamicFieldRenderer";

describe("DynamicFieldRenderer QC_CHECK", () => {
  afterEach(cleanup);
  it("emits the stable composite value contract", () => {
    const onChange = vi.fn();
    render(<DynamicFieldRenderer field={{ id: 1, field_key: "surface", field_type: "QC_CHECK", label_fa: "کیفیت سطح", is_required: true }} value={{}} onChange={onChange} onUpload={vi.fn()} />);
    fireEvent.change(screen.getByRole("combobox"), { target: { value: "FAIL" } });
    expect(onChange).toHaveBeenLastCalledWith({ result: "FAIL" });
  });

  it("keeps an existing file reference when another property changes", () => {
    const onChange = vi.fn();
    render(<DynamicFieldRenderer field={{ id: 1, field_key: "surface", field_type: "QC_CHECK", label_fa: "کیفیت سطح" }} value={{ result: "FAIL", fileId: "file-1" }} onChange={onChange} onUpload={vi.fn()} />);
    fireEvent.change(screen.getByPlaceholderText("یادداشت کنترل کیفیت"), { target: { value: "نیازمند اصلاح" } });
    expect(onChange).toHaveBeenLastCalledWith({ result: "FAIL", fileId: "file-1", note: "نیازمند اصلاح" });
    expect(screen.getByText("مشاهده تصویر")).toHaveAttribute("href", "/api/v1/workflow-files/file-1");
  });
});
