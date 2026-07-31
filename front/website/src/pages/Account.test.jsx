// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import Account from "./Account";

vi.mock("../lib/api",()=>({fetchJSON:vi.fn(path=>{
  if(path==="/api/v1/account/orders") return Promise.resolve({data:[{id:"o1",workflow_instance_id:"w1",workflow_name:"سفارش",order_number:"ORD-1",status:"IN_PROGRESS",created_at:"2026-01-01"}]});
  if(path==="/api/v1/account/proformas") return Promise.resolve({data:[]});
  return Promise.resolve({data:[]});
})}));

describe("customer account",()=>{
  it("shows only the customer's API-provided order",async()=>{
    render(<Account/>);
    expect(await screen.findByText("ORD-1")).toBeInTheDocument();
    expect(screen.getByText("در حال انجام")).toBeInTheDocument();
  });
});
