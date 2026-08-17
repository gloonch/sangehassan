export function moveProduct(products, draggedID, targetID, placement = "before") {
  const fromIndex = products.findIndex((product) => product.id === draggedID);
  const originalTargetIndex = products.findIndex((product) => product.id === targetID);
  if (fromIndex < 0 || originalTargetIndex < 0 || draggedID === targetID) return products;

  const reordered = [...products];
  const [dragged] = reordered.splice(fromIndex, 1);
  let targetIndex = reordered.findIndex((product) => product.id === targetID);
  if (placement === "after") targetIndex += 1;
  reordered.splice(targetIndex, 0, dragged);

  return reordered.map((product, index) => ({ ...product, display_order: index }));
}
