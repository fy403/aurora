(this["webpackJsonp"] = this["webpackJsonp"] || []).push([[12],{

/***/ "./node_modules/@vue/cli-plugin-babel/node_modules/babel-loader/lib/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/views/features/redirect.vue?vue&type=script&lang=js&":
/*!****************************************************************************************************************************************************************************************************************************************************!*\
  !*** ./node_modules/@vue/cli-plugin-babel/node_modules/babel-loader/lib??ref--12-0!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/views/features/redirect.vue?vue&type=script&lang=js& ***!
  \****************************************************************************************************************************************************************************************************************************************************/
/*! exports provided: default */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony default export */ __webpack_exports__["default"] = ({
  created() {
    const {
      params,
      query
    } = this.$route;
    const {
      path
    } = params;
    this.$router.replace({
      path: '/' + path,
      query
    });
  },
  render: function (h) {
    return h(); // avoid warning message
  }
});

/***/ }),

/***/ "./src/views/features/redirect.vue":
/*!*****************************************!*\
  !*** ./src/views/features/redirect.vue ***!
  \*****************************************/
/*! exports provided: default */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony import */ var _redirect_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! ./redirect.vue?vue&type=script&lang=js& */ "./src/views/features/redirect.vue?vue&type=script&lang=js&");
/* empty/unused harmony star reexport *//* harmony import */ var _node_modules_vue_loader_lib_runtime_componentNormalizer_js__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! ../../../node_modules/vue-loader/lib/runtime/componentNormalizer.js */ "./node_modules/vue-loader/lib/runtime/componentNormalizer.js");
var render, staticRenderFns




/* normalize component */

var component = Object(_node_modules_vue_loader_lib_runtime_componentNormalizer_js__WEBPACK_IMPORTED_MODULE_1__["default"])(
  _redirect_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0__["default"],
  render,
  staticRenderFns,
  false,
  null,
  null,
  null
  
)

/* hot reload */
if (true) {
  var api = __webpack_require__(/*! ./node_modules/vue-hot-reload-api/dist/index.js */ "./node_modules/vue-hot-reload-api/dist/index.js")
  api.install(__webpack_require__(/*! vue */ "./node_modules/vue/dist/vue.runtime.esm.js"))
  if (api.compatible) {
    module.hot.accept()
    if (!api.isRecorded('67687bc4')) {
      api.createRecord('67687bc4', component.options)
    } else {
      api.reload('67687bc4', component.options)
    }
    
  }
}
component.options.__file = "src/views/features/redirect.vue"
/* harmony default export */ __webpack_exports__["default"] = (component.exports);

/***/ }),

/***/ "./src/views/features/redirect.vue?vue&type=script&lang=js&":
/*!******************************************************************!*\
  !*** ./src/views/features/redirect.vue?vue&type=script&lang=js& ***!
  \******************************************************************/
/*! exports provided: default */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony import */ var _node_modules_vue_cli_plugin_babel_node_modules_babel_loader_lib_index_js_ref_12_0_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_redirect_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! -!../../../node_modules/@vue/cli-plugin-babel/node_modules/babel-loader/lib??ref--12-0!../../../node_modules/cache-loader/dist/cjs.js??ref--0-0!../../../node_modules/vue-loader/lib??vue-loader-options!./redirect.vue?vue&type=script&lang=js& */ "./node_modules/@vue/cli-plugin-babel/node_modules/babel-loader/lib/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/views/features/redirect.vue?vue&type=script&lang=js&");
/* empty/unused harmony star reexport */ /* harmony default export */ __webpack_exports__["default"] = (_node_modules_vue_cli_plugin_babel_node_modules_babel_loader_lib_index_js_ref_12_0_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_redirect_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0__["default"]); 

/***/ })

}]);
//# sourceMappingURL=12.js.map