(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[853],{55062:function(e,t,r){Promise.resolve().then(r.bind(r,33992))},33992:function(e,t,r){"use strict";r.r(t),r.d(t,{default:function(){return U}});var s=r(57437),n=r(2265),a=r(81628),i=r(23611),o=r(61396),d=r.n(o),l=r(38110),c=r(61865),u=r(74578);let f=n.forwardRef((e,t)=>{let{className:r,type:n,...i}=e;return(0,s.jsx)("input",{type:n,className:(0,a.cn)("flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",r),ref:t,...i})});f.displayName="Input";var m=r(22147),x=r(78043),h=r(91936);let p=x.fC,g=x.xz,j=x.h_;x.x8;let b=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)(x.aV,{ref:t,className:(0,a.cn)("fixed inset-0 z-50 bg-background/80 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",r),...n})});b.displayName=x.aV.displayName;let v=n.forwardRef((e,t)=>{let{className:r,children:n,...i}=e;return(0,s.jsxs)(j,{children:[(0,s.jsx)(b,{}),(0,s.jsxs)(x.VY,{ref:t,className:(0,a.cn)("fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%] sm:rounded-lg max-h-[90%] overflow-y-auto",r),...i,children:[n,(0,s.jsxs)(x.x8,{className:"absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground",children:[(0,s.jsx)(h.Z,{className:"h-4 w-4"}),(0,s.jsx)("span",{className:"sr-only",children:"Close"})]})]})]})});v.displayName=x.VY.displayName;let y=e=>{let{className:t,...r}=e;return(0,s.jsx)("div",{className:(0,a.cn)("flex flex-col space-y-1.5 text-center sm:text-left",t),...r})};y.displayName="DialogHeader";let N=e=>{let{className:t,...r}=e;return(0,s.jsx)("div",{className:(0,a.cn)("flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2",t),...r})};N.displayName="DialogFooter";let _=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)(x.Dx,{ref:t,className:(0,a.cn)("text-lg font-semibold leading-none tracking-tight",r),...n})});_.displayName=x.Dx.displayName;let w=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)(x.dk,{ref:t,className:(0,a.cn)("text-sm text-muted-foreground",r),...n})});w.displayName=x.dk.displayName;var k=r(67256),S=r(36743),T=r(96061);let A=(0,T.j)("text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"),R=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)(S.f,{ref:t,className:(0,a.cn)(A(),r),...n})});R.displayName=S.f.displayName;let C=c.RV,I=n.createContext({}),F=e=>{let{...t}=e;return(0,s.jsx)(I.Provider,{value:{name:t.name},children:(0,s.jsx)(c.Qr,{...t})})},O=()=>{let e=n.useContext(I),t=n.useContext(E),{getFieldState:r,formState:s}=(0,c.Gc)(),a=r(e.name,s);if(!e)throw Error("useFormField should be used within <FormField>");let{id:i}=t;return{id:i,name:e.name,formItemId:"".concat(i,"-form-item"),formDescriptionId:"".concat(i,"-form-item-description"),formMessageId:"".concat(i,"-form-item-message"),...a}},E=n.createContext({}),D=n.forwardRef((e,t)=>{let{className:r,...i}=e,o=n.useId();return(0,s.jsx)(E.Provider,{value:{id:o},children:(0,s.jsx)("div",{ref:t,className:(0,a.cn)("space-y-2",r),...i})})});D.displayName="FormItem";let z=n.forwardRef((e,t)=>{let{className:r,...n}=e,{error:i,formItemId:o}=O();return(0,s.jsx)(R,{ref:t,className:(0,a.cn)(i&&"text-destructive",r),htmlFor:o,...n})});z.displayName="FormLabel";let M=n.forwardRef((e,t)=>{let{...r}=e,{error:n,formItemId:a,formDescriptionId:i,formMessageId:o}=O();return(0,s.jsx)(k.g7,{ref:t,id:a,"aria-describedby":n?"".concat(i," ").concat(o):"".concat(i),"aria-invalid":!!n,...r})});M.displayName="FormControl";let V=n.forwardRef((e,t)=>{let{className:r,...n}=e,{formDescriptionId:i}=O();return(0,s.jsx)("p",{ref:t,id:i,className:(0,a.cn)("text-sm text-muted-foreground",r),...n})});V.displayName="FormDescription";let Z=n.forwardRef((e,t)=>{let{className:r,children:n,...i}=e,{error:o,formMessageId:d}=O(),l=o?String(null==o?void 0:o.message):n;return l?(0,s.jsx)("p",{ref:t,id:d,className:(0,a.cn)("text-sm font-medium text-destructive",r),...i,children:l}):null});Z.displayName="FormMessage";var H=r(71271),P=r(51872);let B=u.Ry({address:u.Z_(),content_byte:u.Z_(),genesis_address:u.Z_(),id:u.Z_()});function J(e){let{setBlock:t}=e,[r,o]=(0,n.useState)(!1),[d,u]=(0,n.useState)((0,P.Z)()),x=(0,c.cI)({resolver:(0,l.F)(B),defaultValues:{address:void 0,content_byte:void 0,genesis_address:void 0,id:void 0},values:{id:d,address:"",content_byte:"",genesis_address:""}});return(0,s.jsxs)(p,{open:r,onOpenChange:o,children:[(0,s.jsx)(g,{asChild:!0,children:(0,s.jsx)(i.z,{variant:"outline",children:"Add"})}),(0,s.jsx)(v,{children:(0,s.jsx)(C,{...x,children:(0,s.jsxs)("form",{className:"space-y-6",children:[(0,s.jsx)(y,{children:(0,s.jsx)(_,{children:"Add Inscription"})}),(0,s.jsx)(F,{control:x.control,name:"id",render:e=>{let{field:t}=e;return(0,s.jsxs)(D,{children:[(0,s.jsx)(z,{children:"id"}),(0,s.jsx)(M,{children:(0,s.jsx)(f,{placeholder:"id",...t})}),(0,s.jsx)(Z,{})]})}}),(0,s.jsx)(F,{control:x.control,name:"address",render:e=>{let{field:t}=e;return(0,s.jsxs)(D,{children:[(0,s.jsx)(z,{children:"address"}),(0,s.jsx)(M,{children:(0,s.jsx)(f,{placeholder:"address",...t})}),(0,s.jsx)(Z,{})]})}}),(0,s.jsx)(F,{control:x.control,name:"genesis_address",render:e=>{let{field:t}=e;return(0,s.jsxs)(D,{children:[(0,s.jsx)(z,{children:"genesis address"}),(0,s.jsx)(M,{children:(0,s.jsx)(f,{placeholder:"genesis_address",...t})}),(0,s.jsx)(Z,{})]})}}),(0,s.jsx)(F,{control:x.control,name:"content_byte",render:e=>{let{field:t}=e;return(0,s.jsxs)(D,{children:[(0,s.jsx)(z,{children:"content byte"}),(0,s.jsx)(M,{children:(0,s.jsx)(m.g,{placeholder:"content_byte",className:"resize-none",rows:8,...t})}),(0,s.jsx)(Z,{})]})}}),(0,s.jsx)(N,{children:(0,s.jsx)(i.z,{onClick:x.handleSubmit(function(e){let{address:r,content_byte:s,genesis_address:n,id:i}=e||{};!function(e){fetch("".concat(a.FH,"/mrc20/JoinMockInscription"),{headers:{"Content-Type":"application/json"},method:"POST",body:JSON.stringify(e)}).then(e=>(0,a.nk)(e)).then(e=>{let{error:r}=e||{};r?(0,H.Am)({title:"Error",description:r,variant:"destructive"}):((0,H.Am)({title:"Succes"}),t(e),o(!1),u((0,P.Z)()),x.reset())}).catch(e=>{(0,H.Am)({title:"Error",description:e.message,variant:"destructive"})})}({address:r,content_byte:btoa(s),content_length:0,content_type:"application/json",curse_type:"string",genesis_address:n,genesis_block_hash:"string",genesis_block_height:0,genesis_fee:"string",genesis_timestamp:0,genesis_tx_id:"string",id:i,location:"string",mime_type:"string",number:0,offset:"string",output:"string",recursion_refs:"string",recursive:!0,sat_coinbase_height:0,sat_ordinal:"string",sat_rarity:"string",timestamp:0,tx_id:"string",value:"string"})}),children:"Save changes"})})]})})})]})}let G=u.Ry({from:u.Z_(),to:u.Z_(),id:u.Z_()});function W(e){let{setBlock:t}=e,[r,o]=(0,n.useState)(!1),[d,u]=(0,n.useState)((0,P.Z)()),m=(0,c.cI)({resolver:(0,l.F)(G),defaultValues:{from:"",to:"",id:""},values:{from:"",to:"",id:d}});return(0,s.jsxs)(p,{open:r,onOpenChange:o,children:[(0,s.jsx)(g,{asChild:!0,children:(0,s.jsx)(i.z,{variant:"outline",children:"Add"})}),(0,s.jsx)(v,{className:"",children:(0,s.jsx)(C,{...m,children:(0,s.jsxs)("form",{className:"space-y-6",children:[(0,s.jsx)(y,{children:(0,s.jsx)(_,{children:"Add Transfer"})}),(0,s.jsx)(F,{control:m.control,name:"id",render:e=>{let{field:t}=e;return(0,s.jsxs)(D,{children:[(0,s.jsx)(z,{children:"id"}),(0,s.jsx)(M,{children:(0,s.jsx)(f,{placeholder:"id",...t})}),(0,s.jsx)(Z,{})]})}}),(0,s.jsx)(F,{control:m.control,name:"from",render:e=>{let{field:t}=e;return(0,s.jsxs)(D,{children:[(0,s.jsx)(z,{children:"from address"}),(0,s.jsx)(M,{children:(0,s.jsx)(f,{placeholder:"from address",...t})}),(0,s.jsx)(Z,{})]})}}),(0,s.jsx)(F,{control:m.control,name:"to",render:e=>{let{field:t}=e;return(0,s.jsxs)(D,{children:[(0,s.jsx)(z,{children:"to address"}),(0,s.jsx)(M,{children:(0,s.jsx)(f,{placeholder:"to address",...t})}),(0,s.jsx)(Z,{})]})}}),(0,s.jsx)(N,{children:(0,s.jsx)(i.z,{onClick:m.handleSubmit(function(e){let{from:r,to:s,id:n}=e||{};!function(e){fetch("".concat(a.FH,"/mrc20/JoinMockTransfer"),{headers:{"Content-Type":"application/json"},method:"POST",body:JSON.stringify(e)}).then(e=>(0,a.nk)(e)).then(e=>{let{error:r}=e||{};r?(0,H.Am)({title:"Error",description:r,variant:"destructive"}):((0,H.Am)({title:"Succes"}),t(e),o(!1),u((0,P.Z)()),m.reset())}).catch(e=>{(0,H.Am)({title:"Error",description:e.message,variant:"destructive"})})}({from:{address:r,block_hash:"string",block_height:0,location:"string",offset:"string",output:"string",timestamp:0,tx_id:"string",value:"string"},id:n,number:0,to:{address:s,block_hash:"string",block_height:0,location:"string",offset:"string",output:"string",timestamp:0,tx_id:"string",value:"string"}})}),children:"Save changes"})})]})})})]})}var L=r(66005);function U(){let[e,t]=(0,n.useState)(!1),[r,o]=(0,n.useState)({error:"Loading..."}),{error:l}=r||{};function c(){fetch("".concat(a.FH,"/mrc20/GetMockBlock"),{headers:{"Content-Type":"application/json"},method:"POST"}).then(e=>(0,a.nk)(e)).then(e=>{o(e)}).catch(e=>{o({error:e.message})})}return(0,n.useEffect)(()=>{e||t(!0)},[e]),(0,n.useEffect)(()=>{e&&c()},[e]),(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)("header",{className:"border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60",children:(0,s.jsxs)("div",{className:"container mx-auto px-4 flex items-center justify-between gap-4 h-14",children:[(0,s.jsx)("h3",{className:"text-lg font-bold",children:"Create new block"}),(0,s.jsx)(d(),{href:"/",children:(0,s.jsx)(i.z,{variant:"outline",children:"Home"})})]})}),(0,s.jsx)("main",{className:"container mx-auto flex flex-col gap-4",children:l?(0,s.jsx)("div",{className:"p-4 text-center whitespace-pre",children:l}):(0,s.jsxs)(s.Fragment,{children:[(0,s.jsxs)("div",{className:"p-4 flex items-center justify-between gap-4",children:[(0,s.jsxs)("div",{className:"break-all",children:[(0,s.jsxs)("h2",{className:"text-2xl font-semibold",children:["Block ",r.block_height]}),(0,s.jsxs)("p",{className:"w-full",children:["hash: ",r.block_hash]})]}),(0,s.jsx)(i.z,{onClick:function(){fetch("".concat(a.FH,"/mrc20/WriteMockBlock"),{headers:{"Content-Type":"application/json"},method:"POST"}).then(e=>(0,a.nk)(e)).then(e=>{let{error:t}=e||{};t?(0,H.Am)({title:"Error",description:t,variant:"destructive"}):((0,H.Am)({title:"Succes"}),c())}).catch(e=>{(0,H.Am)({title:"Error",description:e.message,variant:"destructive"})})},variant:"outline",children:"Write Block"})]}),(0,s.jsxs)("div",{className:"border rounded",children:[(0,s.jsxs)("div",{className:"flex items-center justify-between gap-4 px-4 py-2 border-b",children:[(0,s.jsx)("h3",{className:"text-lg font-medium",children:"Inscriptions"}),(0,s.jsx)(J,{setBlock:o})]}),Array.isArray(r.Inscriptions)?(0,s.jsx)("div",{className:"p-4 grid grid-cols-4 gap-4",children:r.Inscriptions.map(e=>{let{id:t,content_byte:r}=e;return(0,s.jsx)("div",{className:"whitespace-pre aspect-square overflow-hidden flex items-center justify-center rounded border",children:(0,a.Vy)(r)},t)})}):(0,s.jsx)("div",{className:"p-4 text-center",children:"No Data"})]}),(0,s.jsxs)("div",{className:"border rounded",children:[(0,s.jsxs)("div",{className:"flex items-center justify-between gap-4 px-4 py-2 border-b",children:[(0,s.jsx)("h3",{className:"text-lg font-medium",children:"Transfers"}),(0,s.jsx)(W,{setBlock:o})]}),Array.isArray(r.Transfers)?(0,s.jsxs)(L.iA,{children:[(0,s.jsx)(L.xD,{children:(0,s.jsxs)(L.SC,{children:[(0,s.jsx)(L.ss,{children:"ID"}),(0,s.jsx)(L.ss,{children:"From"}),(0,s.jsx)(L.ss,{children:"To"})]})}),(0,s.jsx)(L.RM,{children:r.Transfers.map(e=>(0,s.jsxs)(L.SC,{children:[(0,s.jsx)(L.pj,{children:e.id}),(0,s.jsx)(L.pj,{children:(0,s.jsx)(d(),{href:"/account?address=".concat(e.from.address),className:"underline underline-offset-4",children:e.from.address})}),(0,s.jsx)(L.pj,{children:(0,s.jsx)(d(),{href:"/account?address=".concat(e.to.address),className:"underline underline-offset-4",children:e.to.address})})]},e.id))})]}):(0,s.jsx)("div",{className:"p-4 text-center",children:"No Data"})]})]})})]})}},23611:function(e,t,r){"use strict";r.d(t,{z:function(){return l}});var s=r(57437),n=r(2265),a=r(67256),i=r(96061),o=r(81628);let d=(0,i.j)("inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",{variants:{variant:{default:"bg-primary text-primary-foreground hover:bg-primary/90",destructive:"bg-destructive text-destructive-foreground hover:bg-destructive/90",outline:"border border-input bg-background hover:bg-accent hover:text-accent-foreground",secondary:"bg-secondary text-secondary-foreground hover:bg-secondary/80",ghost:"hover:bg-accent hover:text-accent-foreground",link:"text-primary underline-offset-4 hover:underline"},size:{default:"h-10 px-4 py-2",sm:"h-9 rounded-md px-3",lg:"h-11 rounded-md px-8",icon:"h-10 w-10"}},defaultVariants:{variant:"default",size:"default"}}),l=n.forwardRef((e,t)=>{let{className:r,variant:n,size:i,asChild:l=!1,...c}=e,u=l?a.g7:"button";return(0,s.jsx)(u,{className:(0,o.cn)(d({variant:n,size:i,className:r})),ref:t,...c})});l.displayName="Button"},66005:function(e,t,r){"use strict";r.d(t,{RM:function(){return d},SC:function(){return c},iA:function(){return i},pj:function(){return f},ss:function(){return u},xD:function(){return o}});var s=r(57437),n=r(2265),a=r(81628);let i=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)("div",{className:"relative w-full overflow-auto",children:(0,s.jsx)("table",{ref:t,className:(0,a.cn)("w-full caption-bottom text-sm",r),...n})})});i.displayName="Table";let o=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)("thead",{ref:t,className:(0,a.cn)("[&_tr]:border-b",r),...n})});o.displayName="TableHeader";let d=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)("tbody",{ref:t,className:(0,a.cn)("[&_tr:last-child]:border-0",r),...n})});d.displayName="TableBody";let l=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)("tfoot",{ref:t,className:(0,a.cn)("border-t bg-muted/50 font-medium [&>tr]:last:border-b-0",r),...n})});l.displayName="TableFooter";let c=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)("tr",{ref:t,className:(0,a.cn)("border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted",r),...n})});c.displayName="TableRow";let u=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)("th",{ref:t,className:(0,a.cn)("h-12 px-4 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0",r),...n})});u.displayName="TableHead";let f=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)("td",{ref:t,className:(0,a.cn)("p-4 align-middle [&:has([role=checkbox])]:pr-0",r),...n})});f.displayName="TableCell";let m=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)("caption",{ref:t,className:(0,a.cn)("mt-4 text-sm text-muted-foreground",r),...n})});m.displayName="TableCaption"},22147:function(e,t,r){"use strict";r.d(t,{g:function(){return i}});var s=r(57437),n=r(2265),a=r(81628);let i=n.forwardRef((e,t)=>{let{className:r,...n}=e;return(0,s.jsx)("textarea",{className:(0,a.cn)("flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",r),ref:t,...n})});i.displayName="Textarea"},71271:function(e,t,r){"use strict";r.d(t,{Am:function(){return u},pm:function(){return f}});var s=r(2265);let n=0,a=new Map,i=e=>{if(a.has(e))return;let t=setTimeout(()=>{a.delete(e),c({type:"REMOVE_TOAST",toastId:e})},1e6);a.set(e,t)},o=(e,t)=>{switch(t.type){case"ADD_TOAST":return{...e,toasts:[t.toast,...e.toasts].slice(0,1)};case"UPDATE_TOAST":return{...e,toasts:e.toasts.map(e=>e.id===t.toast.id?{...e,...t.toast}:e)};case"DISMISS_TOAST":{let{toastId:r}=t;return r?i(r):e.toasts.forEach(e=>{i(e.id)}),{...e,toasts:e.toasts.map(e=>e.id===r||void 0===r?{...e,open:!1}:e)}}case"REMOVE_TOAST":if(void 0===t.toastId)return{...e,toasts:[]};return{...e,toasts:e.toasts.filter(e=>e.id!==t.toastId)}}},d=[],l={toasts:[]};function c(e){l=o(l,e),d.forEach(e=>{e(l)})}function u(e){let{...t}=e,r=(n=(n+1)%Number.MAX_SAFE_INTEGER).toString(),s=()=>c({type:"DISMISS_TOAST",toastId:r});return c({type:"ADD_TOAST",toast:{...t,id:r,open:!0,onOpenChange:e=>{e||s()}}}),{id:r,dismiss:s,update:e=>c({type:"UPDATE_TOAST",toast:{...e,id:r}})}}function f(){let[e,t]=s.useState(l);return s.useEffect(()=>(d.push(t),()=>{let e=d.indexOf(t);e>-1&&d.splice(e,1)}),[e]),{...e,toast:u,dismiss:e=>c({type:"DISMISS_TOAST",toastId:e})}}},81628:function(e,t,r){"use strict";r.d(t,{FH:function(){return i},Vy:function(){return d},cn:function(){return a},nk:function(){return o}});var s=r(57042),n=r(74769);function a(){for(var e=arguments.length,t=Array(e),r=0;r<e;r++)t[r]=arguments[r];return(0,n.m6)((0,s.W)(t))}let i="/api/v1";async function o(e){if(!e.ok)try{let t=await e.json(),{error:r}=t||{};return{error:r||"Failed to fetch!"}}catch(e){return{error:"Failed to fetch!"}}return e.json()}let d=e=>{try{return JSON.stringify(JSON.parse(atob(e)),null,2)}catch(t){return e}}}},function(e){e.O(0,[412,396,548,813,971,472,744],function(){return e(e.s=55062)}),_N_E=e.O()}]);