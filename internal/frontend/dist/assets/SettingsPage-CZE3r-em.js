import{c as s,u as y,a as u,r as j,j as e,C as i,b as l,d,e as o,f as r,B as n}from"./index-C5-ONvph.js";/**
 * @license lucide-react v0.562.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const k=[["rect",{width:"20",height:"14",x:"2",y:"3",rx:"2",key:"48i651"}],["line",{x1:"8",x2:"16",y1:"21",y2:"21",key:"1svkeh"}],["line",{x1:"12",x2:"12",y1:"17",y2:"21",key:"vw1qmm"}]],f=s("monitor",k);/**
 * @license lucide-react v0.562.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const g=[["path",{d:"M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401",key:"kfwtm"}]],v=s("moon",g);/**
 * @license lucide-react v0.562.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const b=[["path",{d:"M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8",key:"1357e3"}],["path",{d:"M3 3v5h5",key:"1xhq8a"}]],N=s("rotate-ccw",b);/**
 * @license lucide-react v0.562.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const w=[["circle",{cx:"12",cy:"12",r:"4",key:"4exip2"}],["path",{d:"M12 2v2",key:"tus03m"}],["path",{d:"M12 20v2",key:"1lh1kg"}],["path",{d:"m4.93 4.93 1.41 1.41",key:"149t6j"}],["path",{d:"m17.66 17.66 1.41 1.41",key:"ptbguv"}],["path",{d:"M2 12h2",key:"1t8f8n"}],["path",{d:"M20 12h2",key:"1q8mjw"}],["path",{d:"m6.34 17.66-1.41 1.41",key:"1m8zz5"}],["path",{d:"m19.07 4.93-1.41 1.41",key:"1shlcs"}]],C=s("sun",w);function S(){return y()}const M=[{value:"light",icon:C,label:"Light",description:"Light theme"},{value:"dark",icon:v,label:"Dark",description:"Dark theme"},{value:"system",icon:f,label:"System",description:"Follow system preference"}];function _(){const{theme:a,setTheme:m}=S(),{resetSetup:x}=u(),[h,c]=j.useState(!1),p=()=>{x(),window.location.reload()};return e.jsxs("div",{className:"max-w-2xl space-y-6",children:[e.jsxs("div",{children:[e.jsx("h2",{className:"text-lg font-semibold text-text-primary",children:"Settings"}),e.jsx("p",{className:"text-sm text-text-secondary",children:"Configure your Aether WebUI preferences"})]}),e.jsxs(i,{children:[e.jsx(l,{children:e.jsx(d,{children:"Appearance"})}),e.jsx(o,{children:e.jsxs("div",{className:"space-y-4",children:[e.jsx("label",{className:"text-sm font-medium text-text-primary",children:"Theme"}),e.jsx("div",{className:"grid grid-cols-3 gap-3",children:M.map(t=>e.jsxs("button",{onClick:()=>m(t.value),className:r("flex flex-col items-center gap-2 p-4 rounded-lg border transition-colors",a===t.value?"border-primary-500 bg-primary-50 dark:bg-primary-900/20":"border-border-primary hover:border-border-secondary"),children:[e.jsx(t.icon,{className:r("w-6 h-6",a===t.value?"text-primary-600":"text-text-secondary")}),e.jsx("span",{className:r("text-sm font-medium",a===t.value?"text-primary-600":"text-text-primary"),children:t.label})]},t.value))})]})})]}),e.jsxs(i,{className:"border-red-200 dark:border-red-900",children:[e.jsx(l,{children:e.jsx(d,{className:"text-red-600",children:"Danger Zone"})}),e.jsx(o,{children:e.jsxs("div",{className:"flex items-center justify-between",children:[e.jsxs("div",{children:[e.jsx("p",{className:"font-medium text-text-primary",children:"Run Setup Again"}),e.jsx("p",{className:"text-sm text-text-secondary",children:"Re-run the initial setup wizard. This will not delete any existing data."})]}),h?e.jsxs("div",{className:"flex gap-2",children:[e.jsx(n,{variant:"ghost",onClick:()=>c(!1),children:"Cancel"}),e.jsx(n,{variant:"danger",onClick:p,children:"Confirm"})]}):e.jsx(n,{variant:"secondary",leftIcon:e.jsx(N,{className:"w-4 h-4"}),onClick:()=>c(!0),children:"Reset Setup"})]})})]})]})}export{_ as default};
