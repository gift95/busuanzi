!function(){var t=["site_pv","site_uv","page_pv","page_uv"],e=document.currentScript,a=e.hasAttribute("pjax"),n=e.getAttribute("data-api")||"http://127.0.0.1:8080/api",r=e.getAttribute("data-prefix")||"busuanzi",i=e.getAttribute("data-style")||"default",o="bsz-id",s=function(){var e=new XMLHttpRequest;e.open("POST",n,!0);var a=localStorage.getItem(o);null!=a&&e.setRequestHeader("Authorization","Bearer "+a),e.setRequestHeader("Content-Type","application/json"),e.setRequestHeader("x-bsz-referer",window.location.href),e.onreadystatechange=function(){if(4===e.readyState&&200===e.status){var a=JSON.parse(e.responseText);if(!0===a.success){t.map((function(t){var e=document.getElementById("".concat(r,"_").concat(t));null!=e&&(e.innerHTML=function(t,e){switch(void 0===e&&(e="default"),e){case"comma":return t.toLocaleString();case"short":for(var a=["","K","M","B","T"],n=0;t>=1e3&&n<a.length-1;)t/=1e3,n++;return Math.round(100*t)/100+a[n];default:return t.toString()}}(a.data[t],i));var n=document.getElementById("".concat(r,"_container_").concat(t));null!=n&&(n.style.display="inline")}));var n=e.getResponseHeader("Set-Bsz-Identity");null!=n&&""!=n&&localStorage.setItem(o,n)}}},e.send()};if(s(),a){var u=window.history.pushState;window.history.pushState=function(){u.apply(this,arguments),s()},window.addEventListener("popstate",(function(t){s()}),!1)}}();