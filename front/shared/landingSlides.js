import blockImage01 from "./assets/landing_page/blocks/block-slide-01.webp";
import blockImage01Mobile from "./assets/landing_page/blocks/block-slide-01-mobile.webp";
import blockImage02 from "./assets/landing_page/blocks/block-slide-02.webp";
import blockImage02Mobile from "./assets/landing_page/blocks/block-slide-02-mobile.webp";
import blockImage03 from "./assets/landing_page/blocks/block-slide-03.webp";
import blockImage03Mobile from "./assets/landing_page/blocks/block-slide-03-mobile.webp";
import blockImage04 from "./assets/landing_page/blocks/block-slide-04.webp";
import blockImage04Mobile from "./assets/landing_page/blocks/block-slide-04-mobile.webp";
import blockImage05 from "./assets/landing_page/blocks/block-slide-05.webp";
import blockImage05Mobile from "./assets/landing_page/blocks/block-slide-05-mobile.webp";
import blockImage06 from "./assets/landing_page/blocks/block-slide-06.webp";
import blockImage06Mobile from "./assets/landing_page/blocks/block-slide-06-mobile.webp";
import blockImage07 from "./assets/landing_page/blocks/block-slide-07.webp";
import blockImage07Mobile from "./assets/landing_page/blocks/block-slide-07-mobile.webp";
import blockImage08 from "./assets/landing_page/blocks/block-slide-08.webp";
import blockImage08Mobile from "./assets/landing_page/blocks/block-slide-08-mobile.webp";
import finishesImage01 from "./assets/landing_page/products/finish-slide-01.webp";
import finishesImage01Mobile from "./assets/landing_page/products/finish-slide-01-mobile.webp";
import finishesImage02 from "./assets/landing_page/products/finish-slide-02.webp";
import finishesImage02Mobile from "./assets/landing_page/products/finish-slide-02-mobile.webp";
import finishesImage03 from "./assets/landing_page/products/finish-slide-03.webp";
import finishesImage03Mobile from "./assets/landing_page/products/finish-slide-03-mobile.webp";
import finishesImage04 from "./assets/landing_page/products/finish-slide-04.webp";
import finishesImage04Mobile from "./assets/landing_page/products/finish-slide-04-mobile.webp";
import productImage01 from "./assets/landing_page/products/product-slide-01.webp";
import productImage01Mobile from "./assets/landing_page/products/product-slide-01-mobile.webp";
import productImage02 from "./assets/landing_page/products/product-slide-02.webp";
import productImage02Mobile from "./assets/landing_page/products/product-slide-02-mobile.webp";
import productImage03 from "./assets/landing_page/products/product-slide-03.webp";
import productImage03Mobile from "./assets/landing_page/products/product-slide-03-mobile.webp";

export const defaultLandingDesktopImages = {
  finished: [
    finishesImage01,
    finishesImage02,
    finishesImage03,
    finishesImage04,
    productImage01,
    productImage02,
    productImage03
  ],
  blocks: [
    blockImage01,
    blockImage02,
    blockImage03,
    blockImage04,
    blockImage05,
    blockImage06,
    blockImage07,
    blockImage08
  ]
};

export const defaultLandingSlides = {
  finished: [
    { src: finishesImage01, mobileSrc: finishesImage01Mobile, width: 736, height: 981 },
    { src: finishesImage02, mobileSrc: finishesImage02Mobile, width: 736, height: 1508 },
    { src: finishesImage03, mobileSrc: finishesImage03Mobile, width: 736, height: 1308 },
    { src: finishesImage04, mobileSrc: finishesImage04Mobile, width: 735, height: 825 },
    { src: productImage01, mobileSrc: productImage01Mobile, width: 900, height: 862 },
    { src: productImage02, mobileSrc: productImage02Mobile, width: 900, height: 864 },
    { src: productImage03, mobileSrc: productImage03Mobile, width: 900, height: 850 }
  ],
  blocks: [
    { src: blockImage01, mobileSrc: blockImage01Mobile, width: 900, height: 895 },
    { src: blockImage02, mobileSrc: blockImage02Mobile, width: 900, height: 894 },
    { src: blockImage03, mobileSrc: blockImage03Mobile, width: 725, height: 906 },
    { src: blockImage04, mobileSrc: blockImage04Mobile, width: 900, height: 1123 },
    { src: blockImage05, mobileSrc: blockImage05Mobile, width: 900, height: 1123 },
    { src: blockImage06, mobileSrc: blockImage06Mobile, width: 900, height: 1104 },
    { src: blockImage07, mobileSrc: blockImage07Mobile, width: 900, height: 1106 },
    { src: blockImage08, mobileSrc: blockImage08Mobile, width: 900, height: 1104 }
  ]
};
