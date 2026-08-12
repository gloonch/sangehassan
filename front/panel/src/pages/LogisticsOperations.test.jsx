// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Batches, Shipments } from "./LogisticsOperations";

const fetchJSON = vi.fn();
vi.mock("../lib/api",()=>({fetchJSON:(...args)=>fetchJSON(...args)}));
vi.mock("../lib/auth",()=>({useAuth:()=>({hasPermission:()=>true})}));

describe("logistics operational lists",()=>{
  beforeEach(()=>fetchJSON.mockReset());
  it("renders batch quantities returned as strings",async()=>{
    fetchJSON.mockResolvedValue({data:[{id:"b1",batch_number:"ORD-B01",order_number:"ORD",stone_name:"مرمریت",actual_quantity:"2.5000",planned_quantity:"5.0000",quantity_unit:"TON",status:"IN_PRODUCTION"}]});
    render(<MemoryRouter><Batches/></MemoryRouter>);
    expect(await screen.findByText("ORD-B01")).toBeInTheDocument();
    expect(screen.getByText(/2.5000/)).toBeInTheDocument();
  });
  it("renders only shipment data supplied by the permission-filtered API",async()=>{
    fetchJSON.mockResolvedValue({data:[{id:"s1",shipment_number:"SHP-1",order_number:"ORD",shipment_type:"DOMESTIC_TRUCK",status:"IN_TRANSIT"}]});
    render(<MemoryRouter><Shipments/></MemoryRouter>);
    expect(await screen.findByText("SHP-1")).toBeInTheDocument();
    expect(screen.getByText("در مسیر")).toBeInTheDocument();
  });
});
