// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NotificationBell from "../components/NotificationBell";
import Order from "./Order";

const fetchJSON=vi.fn();
const permissions=new Set();
vi.mock("../lib/api",()=>({fetchJSON:(...args)=>fetchJSON(...args),idempotentHeaders:()=>({"Idempotency-Key":"test-key"})}));
vi.mock("../lib/auth",()=>({useAuth:()=>({hasPermission:code=>permissions.has(code)})}));

describe("finance and documents panel",()=>{
  beforeEach(()=>{fetchJSON.mockReset();permissions.clear()});
  it("shows the unread notification badge only with permission",async()=>{
    permissions.add("notifications.view_own");
    fetchJSON.mockResolvedValue({data:{items:[],unread_count:3}});
    render(<MemoryRouter><NotificationBell/></MemoryRouter>);
    expect(await screen.findByText("3")).toBeInTheDocument();
  });
  it("renders profitability only when the independent permission is present",async()=>{
    permissions.add("finance.commercial_terms.view");permissions.add("finance.profit.view");
    fetchJSON.mockImplementation(path=>path.endsWith("commercial-terms")?Promise.resolve({data:{terms_type:"CUSTOM",currency:"IRR",subtotal:"100",discount_amount:"0",tax_amount:"0",additional_charge_amount:"0",final_customer_amount:"100"}}):Promise.resolve({data:{currency:"IRR",revenue_amount:"100",confirmed_payment_amount:"30",refunded_amount:"0",outstanding_amount:"70",profit_amount:"40"}}));
    render(<MemoryRouter initialEntries={["/dashboard/orders/o1"]}><Routes><Route path="/dashboard/orders/:id" element={<Order/>}/></Routes></MemoryRouter>);
    expect(await screen.findByText("سود تخمینی")).toBeInTheDocument();
    expect(screen.getByText(/۴۰/)).toBeInTheDocument();
  });
});
