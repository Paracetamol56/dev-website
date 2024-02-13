var Module = typeof Module != 'undefined' ? Module : {};
var moduleOverrides = Object.assign({}, Module);
var arguments_ = [];
var thisProgram = './this.program';
var quit_ = (status, toThrow) => {
	throw toThrow;
};
var ENVIRONMENT_IS_WEB = typeof window == 'object';
var ENVIRONMENT_IS_WORKER = typeof importScripts == 'function';
var ENVIRONMENT_IS_NODE =
	typeof process == 'object' &&
	typeof process.versions == 'object' &&
	typeof process.versions.node == 'string';
var ENVIRONMENT_IS_SHELL = !ENVIRONMENT_IS_WEB && !ENVIRONMENT_IS_NODE && !ENVIRONMENT_IS_WORKER;
if (Module['ENVIRONMENT']) {
	throw new Error(
		'Module.ENVIRONMENT has been deprecated. To force the environment, use the ENVIRONMENT compile-time option (for example, -sENVIRONMENT=web or -sENVIRONMENT=node)'
	);
}
var ENVIRONMENT_IS_PTHREAD = Module['ENVIRONMENT_IS_PTHREAD'] || false;
var _scriptDir =
	typeof document != 'undefined' && document.currentScript ? document.currentScript.src : undefined;
if (ENVIRONMENT_IS_WORKER) {
	_scriptDir = self.location.href;
} else if (ENVIRONMENT_IS_NODE) {
	_scriptDir = __filename;
}
var scriptDirectory = '';
function locateFile(path) {
	if (Module['locateFile']) {
		return Module['locateFile'](path, scriptDirectory);
	}
	return scriptDirectory + path;
}
var read_, readAsync, readBinary;
if (ENVIRONMENT_IS_NODE) {
	if (typeof process == 'undefined' || !process.release || process.release.name !== 'node')
		throw new Error(
			'not compiled for this environment (did you build to HTML and try to run it not on the web, or set ENVIRONMENT to something - like node - and run it someplace else - like on the web?)'
		);
	var nodeVersion = process.versions.node;
	var numericVersion = nodeVersion.split('.').slice(0, 3);
	numericVersion =
		numericVersion[0] * 1e4 + numericVersion[1] * 100 + numericVersion[2].split('-')[0] * 1;
	if (numericVersion < 16e4) {
		throw new Error(
			'This emscripten-generated code requires node v16.0.0 (detected v' + nodeVersion + ')'
		);
	}
	var fs = require('fs');
	var nodePath = require('path');
	if (ENVIRONMENT_IS_WORKER) {
		scriptDirectory = nodePath.dirname(scriptDirectory) + '/';
	} else {
		scriptDirectory = __dirname + '/';
	}
	read_ = (filename, binary) => {
		filename = isFileURI(filename) ? new URL(filename) : nodePath.normalize(filename);
		return fs.readFileSync(filename, binary ? undefined : 'utf8');
	};
	readBinary = (filename) => {
		var ret = read_(filename, true);
		if (!ret.buffer) {
			ret = new Uint8Array(ret);
		}
		assert(ret.buffer);
		return ret;
	};
	readAsync = (filename, onload, onerror, binary = true) => {
		filename = isFileURI(filename) ? new URL(filename) : nodePath.normalize(filename);
		fs.readFile(filename, binary ? undefined : 'utf8', (err, data) => {
			if (err) onerror(err);
			else onload(binary ? data.buffer : data);
		});
	};
	if (!Module['thisProgram'] && process.argv.length > 1) {
		thisProgram = process.argv[1].replace(/\\/g, '/');
	}
	arguments_ = process.argv.slice(2);
	if (typeof module != 'undefined') {
		module['exports'] = Module;
	}
	process.on('uncaughtException', (ex) => {
		if (ex !== 'unwind' && !(ex instanceof ExitStatus) && !(ex.context instanceof ExitStatus)) {
			throw ex;
		}
	});
	quit_ = (status, toThrow) => {
		process.exitCode = status;
		throw toThrow;
	};
	Module['inspect'] = () => '[Emscripten Module object]';
	let nodeWorkerThreads;
	try {
		nodeWorkerThreads = require('worker_threads');
	} catch (e) {
		console.error(
			'The "worker_threads" module is not supported in this node.js build - perhaps a newer version is needed?'
		);
		throw e;
	}
	global.Worker = nodeWorkerThreads.Worker;
} else if (ENVIRONMENT_IS_SHELL) {
	if (
		(typeof process == 'object' && typeof require === 'function') ||
		typeof window == 'object' ||
		typeof importScripts == 'function'
	)
		throw new Error(
			'not compiled for this environment (did you build to HTML and try to run it not on the web, or set ENVIRONMENT to something - like node - and run it someplace else - like on the web?)'
		);
	if (typeof read != 'undefined') {
		read_ = read;
	}
	readBinary = (f) => {
		if (typeof readbuffer == 'function') {
			return new Uint8Array(readbuffer(f));
		}
		let data = read(f, 'binary');
		assert(typeof data == 'object');
		return data;
	};
	readAsync = (f, onload, onerror) => {
		setTimeout(() => onload(readBinary(f)));
	};
	if (typeof clearTimeout == 'undefined') {
		globalThis.clearTimeout = (id) => {};
	}
	if (typeof setTimeout == 'undefined') {
		globalThis.setTimeout = (f) => (typeof f == 'function' ? f() : abort());
	}
	if (typeof scriptArgs != 'undefined') {
		arguments_ = scriptArgs;
	} else if (typeof arguments != 'undefined') {
		arguments_ = arguments;
	}
	if (typeof quit == 'function') {
		quit_ = (status, toThrow) => {
			setTimeout(() => {
				if (!(toThrow instanceof ExitStatus)) {
					let toLog = toThrow;
					if (toThrow && typeof toThrow == 'object' && toThrow.stack) {
						toLog = [toThrow, toThrow.stack];
					}
					err(`exiting due to exception: ${toLog}`);
				}
				quit(status);
			});
			throw toThrow;
		};
	}
	if (typeof print != 'undefined') {
		if (typeof console == 'undefined') console = {};
		console.log = print;
		console.warn = console.error = typeof printErr != 'undefined' ? printErr : print;
	}
} else if (ENVIRONMENT_IS_WEB || ENVIRONMENT_IS_WORKER) {
	if (ENVIRONMENT_IS_WORKER) {
		scriptDirectory = self.location.href;
	} else if (typeof document != 'undefined' && document.currentScript) {
		scriptDirectory = document.currentScript.src;
	}
	if (scriptDirectory.indexOf('blob:') !== 0) {
		scriptDirectory = scriptDirectory.substr(
			0,
			scriptDirectory.replace(/[?#].*/, '').lastIndexOf('/') + 1
		);
	} else {
		scriptDirectory = '';
	}
	if (!(typeof window == 'object' || typeof importScripts == 'function'))
		throw new Error(
			'not compiled for this environment (did you build to HTML and try to run it not on the web, or set ENVIRONMENT to something - like node - and run it someplace else - like on the web?)'
		);
	if (!ENVIRONMENT_IS_NODE) {
		read_ = (url) => {
			var xhr = new XMLHttpRequest();
			xhr.open('GET', url, false);
			xhr.send(null);
			return xhr.responseText;
		};
		if (ENVIRONMENT_IS_WORKER) {
			readBinary = (url) => {
				var xhr = new XMLHttpRequest();
				xhr.open('GET', url, false);
				xhr.responseType = 'arraybuffer';
				xhr.send(null);
				return new Uint8Array(xhr.response);
			};
		}
		readAsync = (url, onload, onerror) => {
			var xhr = new XMLHttpRequest();
			xhr.open('GET', url, true);
			xhr.responseType = 'arraybuffer';
			xhr.onload = () => {
				if (xhr.status == 200 || (xhr.status == 0 && xhr.response)) {
					onload(xhr.response);
					return;
				}
				onerror();
			};
			xhr.onerror = onerror;
			xhr.send(null);
		};
	}
} else {
	throw new Error('environment detection error');
}
if (ENVIRONMENT_IS_NODE) {
	if (typeof performance == 'undefined') {
		global.performance = require('perf_hooks').performance;
	}
}
var defaultPrint = console.log.bind(console);
var defaultPrintErr = console.error.bind(console);
if (ENVIRONMENT_IS_NODE) {
	defaultPrint = (...args) => fs.writeSync(1, args.join(' ') + '\n');
	defaultPrintErr = (...args) => fs.writeSync(2, args.join(' ') + '\n');
}
var out = Module['print'] || defaultPrint;
var err = Module['printErr'] || defaultPrintErr;
Object.assign(Module, moduleOverrides);
moduleOverrides = null;
checkIncomingModuleAPI();
if (Module['arguments']) arguments_ = Module['arguments'];
legacyModuleProp('arguments', 'arguments_');
if (Module['thisProgram']) thisProgram = Module['thisProgram'];
legacyModuleProp('thisProgram', 'thisProgram');
if (Module['quit']) quit_ = Module['quit'];
legacyModuleProp('quit', 'quit_');
assert(
	typeof Module['memoryInitializerPrefixURL'] == 'undefined',
	'Module.memoryInitializerPrefixURL option was removed, use Module.locateFile instead'
);
assert(
	typeof Module['pthreadMainPrefixURL'] == 'undefined',
	'Module.pthreadMainPrefixURL option was removed, use Module.locateFile instead'
);
assert(
	typeof Module['cdInitializerPrefixURL'] == 'undefined',
	'Module.cdInitializerPrefixURL option was removed, use Module.locateFile instead'
);
assert(
	typeof Module['filePackagePrefixURL'] == 'undefined',
	'Module.filePackagePrefixURL option was removed, use Module.locateFile instead'
);
assert(typeof Module['read'] == 'undefined', 'Module.read option was removed (modify read_ in JS)');
assert(
	typeof Module['readAsync'] == 'undefined',
	'Module.readAsync option was removed (modify readAsync in JS)'
);
assert(
	typeof Module['readBinary'] == 'undefined',
	'Module.readBinary option was removed (modify readBinary in JS)'
);
assert(
	typeof Module['setWindowTitle'] == 'undefined',
	'Module.setWindowTitle option was removed (modify emscripten_set_window_title in JS)'
);
assert(
	typeof Module['TOTAL_MEMORY'] == 'undefined',
	'Module.TOTAL_MEMORY has been renamed Module.INITIAL_MEMORY'
);
legacyModuleProp('asm', 'wasmExports');
legacyModuleProp('read', 'read_');
legacyModuleProp('readAsync', 'readAsync');
legacyModuleProp('readBinary', 'readBinary');
legacyModuleProp('setWindowTitle', 'setWindowTitle');
assert(
	ENVIRONMENT_IS_WEB || ENVIRONMENT_IS_WORKER || ENVIRONMENT_IS_NODE,
	'Pthreads do not work in this environment yet (need Web Workers, or an alternative to them)'
);
assert(
	!ENVIRONMENT_IS_SHELL,
	"shell environment detected but not enabled at build time.  Add 'shell' to `-sENVIRONMENT` to enable."
);
var wasmBinary;
if (Module['wasmBinary']) wasmBinary = Module['wasmBinary'];
legacyModuleProp('wasmBinary', 'wasmBinary');
if (typeof WebAssembly != 'object') {
	abort('no native wasm support detected');
}
function intArrayFromBase64(s) {
	if (typeof ENVIRONMENT_IS_NODE != 'undefined' && ENVIRONMENT_IS_NODE) {
		var buf = Buffer.from(s, 'base64');
		return new Uint8Array(buf.buffer, buf.byteOffset, buf.length);
	}
	var decoded = atob(s);
	var bytes = new Uint8Array(decoded.length);
	for (var i = 0; i < decoded.length; ++i) {
		bytes[i] = decoded.charCodeAt(i);
	}
	return bytes;
}
var wasmMemory;
var wasmModule;
var ABORT = false;
var EXITSTATUS;
function assert(condition, text) {
	if (!condition) {
		abort('Assertion failed' + (text ? ': ' + text : ''));
	}
}
var HEAP8, HEAPU8, HEAP16, HEAPU16, HEAP32, HEAPU32, HEAPF32, HEAPF64;
function updateMemoryViews() {
	var b = wasmMemory.buffer;
	Module['HEAP8'] = HEAP8 = new Int8Array(b);
	Module['HEAP16'] = HEAP16 = new Int16Array(b);
	Module['HEAPU8'] = HEAPU8 = new Uint8Array(b);
	Module['HEAPU16'] = HEAPU16 = new Uint16Array(b);
	Module['HEAP32'] = HEAP32 = new Int32Array(b);
	Module['HEAPU32'] = HEAPU32 = new Uint32Array(b);
	Module['HEAPF32'] = HEAPF32 = new Float32Array(b);
	Module['HEAPF64'] = HEAPF64 = new Float64Array(b);
}
assert(
	!Module['STACK_SIZE'],
	'STACK_SIZE can no longer be set at runtime.  Use -sSTACK_SIZE at link time'
);
assert(
	typeof Int32Array != 'undefined' &&
		typeof Float64Array !== 'undefined' &&
		Int32Array.prototype.subarray != undefined &&
		Int32Array.prototype.set != undefined,
	'JS engine does not provide full typed array support'
);
var INITIAL_MEMORY = Module['INITIAL_MEMORY'] || 67108864;
legacyModuleProp('INITIAL_MEMORY', 'INITIAL_MEMORY');
assert(
	INITIAL_MEMORY >= 65536,
	'INITIAL_MEMORY should be larger than STACK_SIZE, was ' +
		INITIAL_MEMORY +
		'! (STACK_SIZE=' +
		65536 +
		')'
);
if (ENVIRONMENT_IS_PTHREAD) {
	wasmMemory = Module['wasmMemory'];
} else {
	if (Module['wasmMemory']) {
		wasmMemory = Module['wasmMemory'];
	} else {
		wasmMemory = new WebAssembly.Memory({
			initial: INITIAL_MEMORY / 65536,
			maximum: INITIAL_MEMORY / 65536,
			shared: true
		});
		if (!(wasmMemory.buffer instanceof SharedArrayBuffer)) {
			err(
				'requested a shared WebAssembly.Memory but the returned buffer is not a SharedArrayBuffer, indicating that while the browser has SharedArrayBuffer it does not have WebAssembly threads support - you may need to set a flag'
			);
			if (ENVIRONMENT_IS_NODE) {
				err(
					'(on node you may need: --experimental-wasm-threads --experimental-wasm-bulk-memory and/or recent version)'
				);
			}
			throw Error('bad memory');
		}
	}
}
updateMemoryViews();
INITIAL_MEMORY = wasmMemory.buffer.byteLength;
assert(INITIAL_MEMORY % 65536 === 0);
function writeStackCookie() {
	var max = _emscripten_stack_get_end();
	assert((max & 3) == 0);
	if (max == 0) {
		max += 4;
	}
	HEAPU32[max >> 2] = 34821223;
	HEAPU32[(max + 4) >> 2] = 2310721022;
	HEAPU32[0 >> 2] = 1668509029;
}
function checkStackCookie() {
	if (ABORT) return;
	var max = _emscripten_stack_get_end();
	if (max == 0) {
		max += 4;
	}
	var cookie1 = HEAPU32[max >> 2];
	var cookie2 = HEAPU32[(max + 4) >> 2];
	if (cookie1 != 34821223 || cookie2 != 2310721022) {
		abort(
			`Stack overflow! Stack cookie has been overwritten at ${ptrToString(
				max
			)}, expected hex dwords 0x89BACDFE and 0x2135467, but received ${ptrToString(
				cookie2
			)} ${ptrToString(cookie1)}`
		);
	}
	if (HEAPU32[0 >> 2] != 1668509029) {
		abort('Runtime error: The application has corrupted its heap memory area (address zero)!');
	}
}
(function () {
	var h16 = new Int16Array(1);
	var h8 = new Int8Array(h16.buffer);
	h16[0] = 25459;
	if (h8[0] !== 115 || h8[1] !== 99)
		throw 'Runtime error: expected the system to be little-endian! (Run with -sSUPPORT_BIG_ENDIAN to bypass)';
})();
var __ATPRERUN__ = [];
var __ATINIT__ = [];
var __ATMAIN__ = [];
var __ATEXIT__ = [];
var __ATPOSTRUN__ = [];
var runtimeInitialized = false;
function preRun() {
	assert(!ENVIRONMENT_IS_PTHREAD);
	if (Module['preRun']) {
		if (typeof Module['preRun'] == 'function') Module['preRun'] = [Module['preRun']];
		while (Module['preRun'].length) {
			addOnPreRun(Module['preRun'].shift());
		}
	}
	callRuntimeCallbacks(__ATPRERUN__);
}
function initRuntime() {
	assert(!runtimeInitialized);
	runtimeInitialized = true;
	if (ENVIRONMENT_IS_PTHREAD) return;
	checkStackCookie();
	if (!Module['noFSInit'] && !FS.init.initialized) FS.init();
	FS.ignorePermissions = false;
	TTY.init();
	callRuntimeCallbacks(__ATINIT__);
}
function preMain() {
	checkStackCookie();
	if (ENVIRONMENT_IS_PTHREAD) return;
	callRuntimeCallbacks(__ATMAIN__);
}
function postRun() {
	checkStackCookie();
	if (ENVIRONMENT_IS_PTHREAD) return;
	if (Module['postRun']) {
		if (typeof Module['postRun'] == 'function') Module['postRun'] = [Module['postRun']];
		while (Module['postRun'].length) {
			addOnPostRun(Module['postRun'].shift());
		}
	}
	callRuntimeCallbacks(__ATPOSTRUN__);
}
function addOnPreRun(cb) {
	__ATPRERUN__.unshift(cb);
}
function addOnInit(cb) {
	__ATINIT__.unshift(cb);
}
function addOnPostRun(cb) {
	__ATPOSTRUN__.unshift(cb);
}
assert(
	Math.imul,
	'This browser does not support Math.imul(), build with LEGACY_VM_SUPPORT or POLYFILL_OLD_MATH_FUNCTIONS to add in a polyfill'
);
assert(
	Math.fround,
	'This browser does not support Math.fround(), build with LEGACY_VM_SUPPORT or POLYFILL_OLD_MATH_FUNCTIONS to add in a polyfill'
);
assert(
	Math.clz32,
	'This browser does not support Math.clz32(), build with LEGACY_VM_SUPPORT or POLYFILL_OLD_MATH_FUNCTIONS to add in a polyfill'
);
assert(
	Math.trunc,
	'This browser does not support Math.trunc(), build with LEGACY_VM_SUPPORT or POLYFILL_OLD_MATH_FUNCTIONS to add in a polyfill'
);
var runDependencies = 0;
var runDependencyWatcher = null;
var dependenciesFulfilled = null;
var runDependencyTracking = {};
function getUniqueRunDependency(id) {
	var orig = id;
	while (1) {
		if (!runDependencyTracking[id]) return id;
		id = orig + Math.random();
	}
}
function addRunDependency(id) {
	runDependencies++;
	Module['monitorRunDependencies']?.(runDependencies);
	if (id) {
		assert(!runDependencyTracking[id]);
		runDependencyTracking[id] = 1;
		if (runDependencyWatcher === null && typeof setInterval != 'undefined') {
			runDependencyWatcher = setInterval(() => {
				if (ABORT) {
					clearInterval(runDependencyWatcher);
					runDependencyWatcher = null;
					return;
				}
				var shown = false;
				for (var dep in runDependencyTracking) {
					if (!shown) {
						shown = true;
						err('still waiting on run dependencies:');
					}
					err(`dependency: ${dep}`);
				}
				if (shown) {
					err('(end of list)');
				}
			}, 1e4);
		}
	} else {
		err('warning: run dependency added without ID');
	}
}
function removeRunDependency(id) {
	runDependencies--;
	Module['monitorRunDependencies']?.(runDependencies);
	if (id) {
		assert(runDependencyTracking[id]);
		delete runDependencyTracking[id];
	} else {
		err('warning: run dependency removed without ID');
	}
	if (runDependencies == 0) {
		if (runDependencyWatcher !== null) {
			clearInterval(runDependencyWatcher);
			runDependencyWatcher = null;
		}
		if (dependenciesFulfilled) {
			var callback = dependenciesFulfilled;
			dependenciesFulfilled = null;
			callback();
		}
	}
}
function abort(what) {
	Module['onAbort']?.(what);
	what = 'Aborted(' + what + ')';
	err(what);
	ABORT = true;
	EXITSTATUS = 1;
	if (what.indexOf('RuntimeError: unreachable') >= 0) {
		what +=
			'. "unreachable" may be due to ASYNCIFY_STACK_SIZE not being large enough (try increasing it)';
	}
	var e = new WebAssembly.RuntimeError(what);
	throw e;
}
var dataURIPrefix = 'data:application/octet-stream;base64,';
var isDataURI = (filename) => filename.startsWith(dataURIPrefix);
var isFileURI = (filename) => filename.startsWith('file://');
function createExportWrapper(name) {
	return function () {
		assert(runtimeInitialized, `native function \`${name}\` called before runtime initialization`);
		var f = wasmExports[name];
		assert(f, `exported native function \`${name}\` not found`);
		return f.apply(null, arguments);
	};
}
var wasmBinaryFile;
wasmBinaryFile = 'barnes-hut.wasm';
if (!isDataURI(wasmBinaryFile)) {
	wasmBinaryFile = locateFile(wasmBinaryFile);
}
function getBinarySync(file) {
	if (file == wasmBinaryFile && wasmBinary) {
		return new Uint8Array(wasmBinary);
	}
	if (readBinary) {
		return readBinary(file);
	}
	throw 'both async and sync fetching of the wasm failed';
}
function getBinaryPromise(binaryFile) {
	if (!wasmBinary && (ENVIRONMENT_IS_WEB || ENVIRONMENT_IS_WORKER)) {
		if (typeof fetch == 'function' && !isFileURI(binaryFile)) {
			return fetch(binaryFile, { credentials: 'same-origin' })
				.then((response) => {
					if (!response['ok']) {
						throw "failed to load wasm binary file at '" + binaryFile + "'";
					}
					return response['arrayBuffer']();
				})
				.catch(() => getBinarySync(binaryFile));
		} else if (readAsync) {
			return new Promise((resolve, reject) => {
				readAsync(binaryFile, (response) => resolve(new Uint8Array(response)), reject);
			});
		}
	}
	return Promise.resolve().then(() => getBinarySync(binaryFile));
}
function instantiateArrayBuffer(binaryFile, imports, receiver) {
	return getBinaryPromise(binaryFile)
		.then((binary) => WebAssembly.instantiate(binary, imports))
		.then((instance) => instance)
		.then(receiver, (reason) => {
			err(`failed to asynchronously prepare wasm: ${reason}`);
			if (isFileURI(wasmBinaryFile)) {
				err(
					`warning: Loading from a file URI (${wasmBinaryFile}) is not supported in most browsers. See https://emscripten.org/docs/getting_started/FAQ.html#how-do-i-run-a-local-webserver-for-testing-why-does-my-program-stall-in-downloading-or-preparing`
				);
			}
			abort(reason);
		});
}
function instantiateAsync(binary, binaryFile, imports, callback) {
	if (
		!binary &&
		typeof WebAssembly.instantiateStreaming == 'function' &&
		!isDataURI(binaryFile) &&
		!isFileURI(binaryFile) &&
		!ENVIRONMENT_IS_NODE &&
		typeof fetch == 'function'
	) {
		return fetch(binaryFile, { credentials: 'same-origin' }).then((response) => {
			var result = WebAssembly.instantiateStreaming(response, imports);
			return result.then(callback, function (reason) {
				err(`wasm streaming compile failed: ${reason}`);
				err('falling back to ArrayBuffer instantiation');
				return instantiateArrayBuffer(binaryFile, imports, callback);
			});
		});
	}
	return instantiateArrayBuffer(binaryFile, imports, callback);
}
function createWasm() {
	var info = { env: wasmImports, wasi_snapshot_preview1: wasmImports };
	function receiveInstance(instance, module) {
		wasmExports = instance.exports;
		wasmExports = Asyncify.instrumentWasmExports(wasmExports);
		registerTLSInit(wasmExports['_emscripten_tls_init']);
		addOnInit(wasmExports['__wasm_call_ctors']);
		wasmModule = module;
		removeRunDependency('wasm-instantiate');
		return wasmExports;
	}
	addRunDependency('wasm-instantiate');
	var trueModule = Module;
	function receiveInstantiationResult(result) {
		assert(
			Module === trueModule,
			'the Module object should not be replaced during async compilation - perhaps the order of HTML elements is wrong?'
		);
		trueModule = null;
		receiveInstance(result['instance'], result['module']);
	}
	if (Module['instantiateWasm']) {
		try {
			return Module['instantiateWasm'](info, receiveInstance);
		} catch (e) {
			err(`Module.instantiateWasm callback failed with error: ${e}`);
			return false;
		}
	}
	instantiateAsync(wasmBinary, wasmBinaryFile, info, receiveInstantiationResult);
	return {};
}
var tempDouble;
var tempI64;
function legacyModuleProp(prop, newName, incomming = true) {
	if (!Object.getOwnPropertyDescriptor(Module, prop)) {
		Object.defineProperty(Module, prop, {
			configurable: true,
			get() {
				let extra = incomming
					? ' (the initial value can be provided on Module, but after startup the value is only looked for on a local variable of that name)'
					: '';
				abort(`\`Module.${prop}\` has been replaced by \`${newName}\`` + extra);
			}
		});
	}
}
function ignoredModuleProp(prop) {
	if (Object.getOwnPropertyDescriptor(Module, prop)) {
		abort(`\`Module.${prop}\` was supplied but \`${prop}\` not included in INCOMING_MODULE_JS_API`);
	}
}
function isExportedByForceFilesystem(name) {
	return (
		name === 'FS_createPath' ||
		name === 'FS_createDataFile' ||
		name === 'FS_createPreloadedFile' ||
		name === 'FS_unlink' ||
		name === 'addRunDependency' ||
		name === 'FS_createLazyFile' ||
		name === 'FS_createDevice' ||
		name === 'removeRunDependency'
	);
}
function missingGlobal(sym, msg) {
	if (typeof globalThis !== 'undefined') {
		Object.defineProperty(globalThis, sym, {
			configurable: true,
			get() {
				warnOnce(`\`${sym}\` is not longer defined by emscripten. ${msg}`);
				return undefined;
			}
		});
	}
}
missingGlobal('buffer', 'Please use HEAP8.buffer or wasmMemory.buffer');
missingGlobal('asm', 'Please use wasmExports instead');
function missingLibrarySymbol(sym) {
	if (typeof globalThis !== 'undefined' && !Object.getOwnPropertyDescriptor(globalThis, sym)) {
		Object.defineProperty(globalThis, sym, {
			configurable: true,
			get() {
				var msg = `\`${sym}\` is a library symbol and not included by default; add it to your library.js __deps or to DEFAULT_LIBRARY_FUNCS_TO_INCLUDE on the command line`;
				var librarySymbol = sym;
				if (!librarySymbol.startsWith('_')) {
					librarySymbol = '$' + sym;
				}
				msg += ` (e.g. -sDEFAULT_LIBRARY_FUNCS_TO_INCLUDE='${librarySymbol}')`;
				if (isExportedByForceFilesystem(sym)) {
					msg +=
						'. Alternatively, forcing filesystem support (-sFORCE_FILESYSTEM) can export this for you';
				}
				warnOnce(msg);
				return undefined;
			}
		});
	}
	unexportedRuntimeSymbol(sym);
}
function unexportedRuntimeSymbol(sym) {
	if (!Object.getOwnPropertyDescriptor(Module, sym)) {
		Object.defineProperty(Module, sym, {
			configurable: true,
			get() {
				var msg = `'${sym}' was not exported. add it to EXPORTED_RUNTIME_METHODS (see the Emscripten FAQ)`;
				if (isExportedByForceFilesystem(sym)) {
					msg +=
						'. Alternatively, forcing filesystem support (-sFORCE_FILESYSTEM) can export this for you';
				}
				abort(msg);
			}
		});
	}
}
function dbg(text) {
	if (ENVIRONMENT_IS_NODE) {
		fs.writeSync(2, Array.from(arguments).join(' ') + '\n');
	} else console.warn.apply(console, arguments);
}
var ASM_CONSTS = {
	38352: () => {
		if (document.fullscreenElement) return 1;
	},
	38398: () => document.getElementById('canvas').width,
	38450: () => parseInt(document.getElementById('canvas').style.width),
	38518: () => {
		document.exitFullscreen();
	},
	38545: () => {
		setTimeout(function () {
			Module.requestFullscreen(false, false);
		}, 100);
	},
	38618: () => {
		if (document.fullscreenElement) return 1;
	},
	38664: () => document.getElementById('canvas').width,
	38716: () => screen.width,
	38741: () => {
		document.exitFullscreen();
	},
	38768: () => {
		setTimeout(function () {
			Module.requestFullscreen(false, true);
			setTimeout(function () {
				canvas.style.width = 'unset';
			}, 100);
		}, 100);
	},
	38901: () => {
		if (document.fullscreenElement) return 1;
	},
	38947: () => document.getElementById('canvas').width,
	38999: () => parseInt(document.getElementById('canvas').style.width),
	39067: () => {
		if (document.fullscreenElement) return 1;
	},
	39113: () => document.getElementById('canvas').width,
	39165: () => screen.width,
	39190: () => {
		if (document.fullscreenElement) return 1;
	},
	39236: () => document.getElementById('canvas').width,
	39288: () => screen.width,
	39313: () => {
		document.exitFullscreen();
	},
	39340: () => {
		if (document.fullscreenElement) return 1;
	},
	39386: () => document.getElementById('canvas').width,
	39438: () => parseInt(document.getElementById('canvas').style.width),
	39506: () => {
		document.exitFullscreen();
	},
	39533: () => screen.width,
	39558: () => screen.height,
	39584: () => window.screenX,
	39611: () => window.screenY,
	39638: ($0) => {
		navigator.clipboard.writeText(UTF8ToString($0));
	},
	39691: ($0) => {
		document.getElementById('canvas').style.cursor = UTF8ToString($0);
	},
	39762: () => {
		document.getElementById('canvas').style.cursor = 'none';
	},
	39819: ($0) => {
		document.getElementById('canvas').style.cursor = UTF8ToString($0);
	},
	39890: () => {
		if (document.fullscreenElement) return 1;
	},
	39936: () => {
		if (document.pointerLockElement) return 1;
	}
};
function GetWindowInnerWidth() {
	return window.innerWidth;
}
function GetWindowInnerHeight() {
	return window.innerHeight;
}
function ExitStatus(status) {
	this.name = 'ExitStatus';
	this.message = `Program terminated with exit(${status})`;
	this.status = status;
}
var terminateWorker = (worker) => {
	worker.terminate();
	worker.onmessage = (e) => {
		var cmd = e['data']['cmd'];
		err(`received "${cmd}" command from terminated worker: ${worker.workerID}`);
	};
};
var killThread = (pthread_ptr) => {
	assert(
		!ENVIRONMENT_IS_PTHREAD,
		'Internal Error! killThread() can only ever be called from main application thread!'
	);
	assert(pthread_ptr, 'Internal Error! Null pthread_ptr in killThread!');
	var worker = PThread.pthreads[pthread_ptr];
	delete PThread.pthreads[pthread_ptr];
	terminateWorker(worker);
	__emscripten_thread_free_data(pthread_ptr);
	PThread.runningWorkers.splice(PThread.runningWorkers.indexOf(worker), 1);
	worker.pthread_ptr = 0;
};
var cancelThread = (pthread_ptr) => {
	assert(
		!ENVIRONMENT_IS_PTHREAD,
		'Internal Error! cancelThread() can only ever be called from main application thread!'
	);
	assert(pthread_ptr, 'Internal Error! Null pthread_ptr in cancelThread!');
	var worker = PThread.pthreads[pthread_ptr];
	worker.postMessage({ cmd: 'cancel' });
};
var cleanupThread = (pthread_ptr) => {
	assert(
		!ENVIRONMENT_IS_PTHREAD,
		'Internal Error! cleanupThread() can only ever be called from main application thread!'
	);
	assert(pthread_ptr, 'Internal Error! Null pthread_ptr in cleanupThread!');
	var worker = PThread.pthreads[pthread_ptr];
	assert(worker);
	PThread.returnWorkerToPool(worker);
};
var spawnThread = (threadParams) => {
	assert(
		!ENVIRONMENT_IS_PTHREAD,
		'Internal Error! spawnThread() can only ever be called from main application thread!'
	);
	assert(threadParams.pthread_ptr, 'Internal error, no pthread ptr!');
	var worker = PThread.getNewWorker();
	if (!worker) {
		return 6;
	}
	assert(!worker.pthread_ptr, 'Internal error!');
	PThread.runningWorkers.push(worker);
	PThread.pthreads[threadParams.pthread_ptr] = worker;
	worker.pthread_ptr = threadParams.pthread_ptr;
	var msg = {
		cmd: 'run',
		start_routine: threadParams.startRoutine,
		arg: threadParams.arg,
		pthread_ptr: threadParams.pthread_ptr
	};
	if (ENVIRONMENT_IS_NODE) {
		worker.unref();
	}
	worker.postMessage(msg, threadParams.transferList);
	return 0;
};
var runtimeKeepaliveCounter = 0;
var keepRuntimeAlive = () => noExitRuntime || runtimeKeepaliveCounter > 0;
var PATH = {
	isAbs: (path) => path.charAt(0) === '/',
	splitPath: (filename) => {
		var splitPathRe = /^(\/?|)([\s\S]*?)((?:\.{1,2}|[^\/]+?|)(\.[^.\/]*|))(?:[\/]*)$/;
		return splitPathRe.exec(filename).slice(1);
	},
	normalizeArray: (parts, allowAboveRoot) => {
		var up = 0;
		for (var i = parts.length - 1; i >= 0; i--) {
			var last = parts[i];
			if (last === '.') {
				parts.splice(i, 1);
			} else if (last === '..') {
				parts.splice(i, 1);
				up++;
			} else if (up) {
				parts.splice(i, 1);
				up--;
			}
		}
		if (allowAboveRoot) {
			for (; up; up--) {
				parts.unshift('..');
			}
		}
		return parts;
	},
	normalize: (path) => {
		var isAbsolute = PATH.isAbs(path),
			trailingSlash = path.substr(-1) === '/';
		path = PATH.normalizeArray(
			path.split('/').filter((p) => !!p),
			!isAbsolute
		).join('/');
		if (!path && !isAbsolute) {
			path = '.';
		}
		if (path && trailingSlash) {
			path += '/';
		}
		return (isAbsolute ? '/' : '') + path;
	},
	dirname: (path) => {
		var result = PATH.splitPath(path),
			root = result[0],
			dir = result[1];
		if (!root && !dir) {
			return '.';
		}
		if (dir) {
			dir = dir.substr(0, dir.length - 1);
		}
		return root + dir;
	},
	basename: (path) => {
		if (path === '/') return '/';
		path = PATH.normalize(path);
		path = path.replace(/\/$/, '');
		var lastSlash = path.lastIndexOf('/');
		if (lastSlash === -1) return path;
		return path.substr(lastSlash + 1);
	},
	join: function () {
		var paths = Array.prototype.slice.call(arguments);
		return PATH.normalize(paths.join('/'));
	},
	join2: (l, r) => PATH.normalize(l + '/' + r)
};
var initRandomFill = () => {
	if (typeof crypto == 'object' && typeof crypto['getRandomValues'] == 'function') {
		return (view) => (view.set(crypto.getRandomValues(new Uint8Array(view.byteLength))), view);
	} else if (ENVIRONMENT_IS_NODE) {
		try {
			var crypto_module = require('crypto');
			var randomFillSync = crypto_module['randomFillSync'];
			if (randomFillSync) {
				return (view) => crypto_module['randomFillSync'](view);
			}
			var randomBytes = crypto_module['randomBytes'];
			return (view) => (view.set(randomBytes(view.byteLength)), view);
		} catch (e) {}
	}
	abort(
		'no cryptographic support found for randomDevice. consider polyfilling it if you want to use something insecure like Math.random(), e.g. put this in a --pre-js: var crypto = { getRandomValues: (array) => { for (var i = 0; i < array.length; i++) array[i] = (Math.random()*256)|0 } };'
	);
};
var randomFill = (view) => (randomFill = initRandomFill())(view);
var PATH_FS = {
	resolve: function () {
		var resolvedPath = '',
			resolvedAbsolute = false;
		for (var i = arguments.length - 1; i >= -1 && !resolvedAbsolute; i--) {
			var path = i >= 0 ? arguments[i] : FS.cwd();
			if (typeof path != 'string') {
				throw new TypeError('Arguments to path.resolve must be strings');
			} else if (!path) {
				return '';
			}
			resolvedPath = path + '/' + resolvedPath;
			resolvedAbsolute = PATH.isAbs(path);
		}
		resolvedPath = PATH.normalizeArray(
			resolvedPath.split('/').filter((p) => !!p),
			!resolvedAbsolute
		).join('/');
		return (resolvedAbsolute ? '/' : '') + resolvedPath || '.';
	},
	relative: (from, to) => {
		from = PATH_FS.resolve(from).substr(1);
		to = PATH_FS.resolve(to).substr(1);
		function trim(arr) {
			var start = 0;
			for (; start < arr.length; start++) {
				if (arr[start] !== '') break;
			}
			var end = arr.length - 1;
			for (; end >= 0; end--) {
				if (arr[end] !== '') break;
			}
			if (start > end) return [];
			return arr.slice(start, end - start + 1);
		}
		var fromParts = trim(from.split('/'));
		var toParts = trim(to.split('/'));
		var length = Math.min(fromParts.length, toParts.length);
		var samePartsLength = length;
		for (var i = 0; i < length; i++) {
			if (fromParts[i] !== toParts[i]) {
				samePartsLength = i;
				break;
			}
		}
		var outputParts = [];
		for (var i = samePartsLength; i < fromParts.length; i++) {
			outputParts.push('..');
		}
		outputParts = outputParts.concat(toParts.slice(samePartsLength));
		return outputParts.join('/');
	}
};
var UTF8Decoder = typeof TextDecoder != 'undefined' ? new TextDecoder('utf8') : undefined;
var UTF8ArrayToString = (heapOrArray, idx, maxBytesToRead) => {
	var endIdx = idx + maxBytesToRead;
	var endPtr = idx;
	while (heapOrArray[endPtr] && !(endPtr >= endIdx)) ++endPtr;
	if (endPtr - idx > 16 && heapOrArray.buffer && UTF8Decoder) {
		return UTF8Decoder.decode(
			heapOrArray.buffer instanceof SharedArrayBuffer
				? heapOrArray.slice(idx, endPtr)
				: heapOrArray.subarray(idx, endPtr)
		);
	}
	var str = '';
	while (idx < endPtr) {
		var u0 = heapOrArray[idx++];
		if (!(u0 & 128)) {
			str += String.fromCharCode(u0);
			continue;
		}
		var u1 = heapOrArray[idx++] & 63;
		if ((u0 & 224) == 192) {
			str += String.fromCharCode(((u0 & 31) << 6) | u1);
			continue;
		}
		var u2 = heapOrArray[idx++] & 63;
		if ((u0 & 240) == 224) {
			u0 = ((u0 & 15) << 12) | (u1 << 6) | u2;
		} else {
			if ((u0 & 248) != 240)
				warnOnce(
					'Invalid UTF-8 leading byte ' +
						ptrToString(u0) +
						' encountered when deserializing a UTF-8 string in wasm memory to a JS string!'
				);
			u0 = ((u0 & 7) << 18) | (u1 << 12) | (u2 << 6) | (heapOrArray[idx++] & 63);
		}
		if (u0 < 65536) {
			str += String.fromCharCode(u0);
		} else {
			var ch = u0 - 65536;
			str += String.fromCharCode(55296 | (ch >> 10), 56320 | (ch & 1023));
		}
	}
	return str;
};
var FS_stdin_getChar_buffer = [];
var lengthBytesUTF8 = (str) => {
	var len = 0;
	for (var i = 0; i < str.length; ++i) {
		var c = str.charCodeAt(i);
		if (c <= 127) {
			len++;
		} else if (c <= 2047) {
			len += 2;
		} else if (c >= 55296 && c <= 57343) {
			len += 4;
			++i;
		} else {
			len += 3;
		}
	}
	return len;
};
var stringToUTF8Array = (str, heap, outIdx, maxBytesToWrite) => {
	assert(typeof str === 'string', `stringToUTF8Array expects a string (got ${typeof str})`);
	if (!(maxBytesToWrite > 0)) return 0;
	var startIdx = outIdx;
	var endIdx = outIdx + maxBytesToWrite - 1;
	for (var i = 0; i < str.length; ++i) {
		var u = str.charCodeAt(i);
		if (u >= 55296 && u <= 57343) {
			var u1 = str.charCodeAt(++i);
			u = (65536 + ((u & 1023) << 10)) | (u1 & 1023);
		}
		if (u <= 127) {
			if (outIdx >= endIdx) break;
			heap[outIdx++] = u;
		} else if (u <= 2047) {
			if (outIdx + 1 >= endIdx) break;
			heap[outIdx++] = 192 | (u >> 6);
			heap[outIdx++] = 128 | (u & 63);
		} else if (u <= 65535) {
			if (outIdx + 2 >= endIdx) break;
			heap[outIdx++] = 224 | (u >> 12);
			heap[outIdx++] = 128 | ((u >> 6) & 63);
			heap[outIdx++] = 128 | (u & 63);
		} else {
			if (outIdx + 3 >= endIdx) break;
			if (u > 1114111)
				warnOnce(
					'Invalid Unicode code point ' +
						ptrToString(u) +
						' encountered when serializing a JS string to a UTF-8 string in wasm memory! (Valid unicode code points should be in range 0-0x10FFFF).'
				);
			heap[outIdx++] = 240 | (u >> 18);
			heap[outIdx++] = 128 | ((u >> 12) & 63);
			heap[outIdx++] = 128 | ((u >> 6) & 63);
			heap[outIdx++] = 128 | (u & 63);
		}
	}
	heap[outIdx] = 0;
	return outIdx - startIdx;
};
function intArrayFromString(stringy, dontAddNull, length) {
	var len = length > 0 ? length : lengthBytesUTF8(stringy) + 1;
	var u8array = new Array(len);
	var numBytesWritten = stringToUTF8Array(stringy, u8array, 0, u8array.length);
	if (dontAddNull) u8array.length = numBytesWritten;
	return u8array;
}
var FS_stdin_getChar = () => {
	if (!FS_stdin_getChar_buffer.length) {
		var result = null;
		if (ENVIRONMENT_IS_NODE) {
			var BUFSIZE = 256;
			var buf = Buffer.alloc(BUFSIZE);
			var bytesRead = 0;
			var fd = process.stdin.fd;
			try {
				bytesRead = fs.readSync(fd, buf);
			} catch (e) {
				if (e.toString().includes('EOF')) bytesRead = 0;
				else throw e;
			}
			if (bytesRead > 0) {
				result = buf.slice(0, bytesRead).toString('utf-8');
			} else {
				result = null;
			}
		} else if (typeof window != 'undefined' && typeof window.prompt == 'function') {
			result = window.prompt('Input: ');
			if (result !== null) {
				result += '\n';
			}
		} else if (typeof readline == 'function') {
			result = readline();
			if (result !== null) {
				result += '\n';
			}
		}
		if (!result) {
			return null;
		}
		FS_stdin_getChar_buffer = intArrayFromString(result, true);
	}
	return FS_stdin_getChar_buffer.shift();
};
var TTY = {
	ttys: [],
	init() {},
	shutdown() {},
	register(dev, ops) {
		TTY.ttys[dev] = { input: [], output: [], ops: ops };
		FS.registerDevice(dev, TTY.stream_ops);
	},
	stream_ops: {
		open(stream) {
			var tty = TTY.ttys[stream.node.rdev];
			if (!tty) {
				throw new FS.ErrnoError(43);
			}
			stream.tty = tty;
			stream.seekable = false;
		},
		close(stream) {
			stream.tty.ops.fsync(stream.tty);
		},
		fsync(stream) {
			stream.tty.ops.fsync(stream.tty);
		},
		read(stream, buffer, offset, length, pos) {
			if (!stream.tty || !stream.tty.ops.get_char) {
				throw new FS.ErrnoError(60);
			}
			var bytesRead = 0;
			for (var i = 0; i < length; i++) {
				var result;
				try {
					result = stream.tty.ops.get_char(stream.tty);
				} catch (e) {
					throw new FS.ErrnoError(29);
				}
				if (result === undefined && bytesRead === 0) {
					throw new FS.ErrnoError(6);
				}
				if (result === null || result === undefined) break;
				bytesRead++;
				buffer[offset + i] = result;
			}
			if (bytesRead) {
				stream.node.timestamp = Date.now();
			}
			return bytesRead;
		},
		write(stream, buffer, offset, length, pos) {
			if (!stream.tty || !stream.tty.ops.put_char) {
				throw new FS.ErrnoError(60);
			}
			try {
				for (var i = 0; i < length; i++) {
					stream.tty.ops.put_char(stream.tty, buffer[offset + i]);
				}
			} catch (e) {
				throw new FS.ErrnoError(29);
			}
			if (length) {
				stream.node.timestamp = Date.now();
			}
			return i;
		}
	},
	default_tty_ops: {
		get_char(tty) {
			return FS_stdin_getChar();
		},
		put_char(tty, val) {
			if (val === null || val === 10) {
				out(UTF8ArrayToString(tty.output, 0));
				tty.output = [];
			} else {
				if (val != 0) tty.output.push(val);
			}
		},
		fsync(tty) {
			if (tty.output && tty.output.length > 0) {
				out(UTF8ArrayToString(tty.output, 0));
				tty.output = [];
			}
		},
		ioctl_tcgets(tty) {
			return {
				c_iflag: 25856,
				c_oflag: 5,
				c_cflag: 191,
				c_lflag: 35387,
				c_cc: [
					3, 28, 127, 21, 4, 0, 1, 0, 17, 19, 26, 0, 18, 15, 23, 22, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0
				]
			};
		},
		ioctl_tcsets(tty, optional_actions, data) {
			return 0;
		},
		ioctl_tiocgwinsz(tty) {
			return [24, 80];
		}
	},
	default_tty1_ops: {
		put_char(tty, val) {
			if (val === null || val === 10) {
				err(UTF8ArrayToString(tty.output, 0));
				tty.output = [];
			} else {
				if (val != 0) tty.output.push(val);
			}
		},
		fsync(tty) {
			if (tty.output && tty.output.length > 0) {
				err(UTF8ArrayToString(tty.output, 0));
				tty.output = [];
			}
		}
	}
};
var mmapAlloc = (size) => {
	abort(
		'internal error: mmapAlloc called but `emscripten_builtin_memalign` native symbol not exported'
	);
};
var MEMFS = {
	ops_table: null,
	mount(mount) {
		return MEMFS.createNode(null, '/', 16384 | 511, 0);
	},
	createNode(parent, name, mode, dev) {
		if (FS.isBlkdev(mode) || FS.isFIFO(mode)) {
			throw new FS.ErrnoError(63);
		}
		MEMFS.ops_table ||= {
			dir: {
				node: {
					getattr: MEMFS.node_ops.getattr,
					setattr: MEMFS.node_ops.setattr,
					lookup: MEMFS.node_ops.lookup,
					mknod: MEMFS.node_ops.mknod,
					rename: MEMFS.node_ops.rename,
					unlink: MEMFS.node_ops.unlink,
					rmdir: MEMFS.node_ops.rmdir,
					readdir: MEMFS.node_ops.readdir,
					symlink: MEMFS.node_ops.symlink
				},
				stream: { llseek: MEMFS.stream_ops.llseek }
			},
			file: {
				node: { getattr: MEMFS.node_ops.getattr, setattr: MEMFS.node_ops.setattr },
				stream: {
					llseek: MEMFS.stream_ops.llseek,
					read: MEMFS.stream_ops.read,
					write: MEMFS.stream_ops.write,
					allocate: MEMFS.stream_ops.allocate,
					mmap: MEMFS.stream_ops.mmap,
					msync: MEMFS.stream_ops.msync
				}
			},
			link: {
				node: {
					getattr: MEMFS.node_ops.getattr,
					setattr: MEMFS.node_ops.setattr,
					readlink: MEMFS.node_ops.readlink
				},
				stream: {}
			},
			chrdev: {
				node: { getattr: MEMFS.node_ops.getattr, setattr: MEMFS.node_ops.setattr },
				stream: FS.chrdev_stream_ops
			}
		};
		var node = FS.createNode(parent, name, mode, dev);
		if (FS.isDir(node.mode)) {
			node.node_ops = MEMFS.ops_table.dir.node;
			node.stream_ops = MEMFS.ops_table.dir.stream;
			node.contents = {};
		} else if (FS.isFile(node.mode)) {
			node.node_ops = MEMFS.ops_table.file.node;
			node.stream_ops = MEMFS.ops_table.file.stream;
			node.usedBytes = 0;
			node.contents = null;
		} else if (FS.isLink(node.mode)) {
			node.node_ops = MEMFS.ops_table.link.node;
			node.stream_ops = MEMFS.ops_table.link.stream;
		} else if (FS.isChrdev(node.mode)) {
			node.node_ops = MEMFS.ops_table.chrdev.node;
			node.stream_ops = MEMFS.ops_table.chrdev.stream;
		}
		node.timestamp = Date.now();
		if (parent) {
			parent.contents[name] = node;
			parent.timestamp = node.timestamp;
		}
		return node;
	},
	getFileDataAsTypedArray(node) {
		if (!node.contents) return new Uint8Array(0);
		if (node.contents.subarray) return node.contents.subarray(0, node.usedBytes);
		return new Uint8Array(node.contents);
	},
	expandFileStorage(node, newCapacity) {
		var prevCapacity = node.contents ? node.contents.length : 0;
		if (prevCapacity >= newCapacity) return;
		var CAPACITY_DOUBLING_MAX = 1024 * 1024;
		newCapacity = Math.max(
			newCapacity,
			(prevCapacity * (prevCapacity < CAPACITY_DOUBLING_MAX ? 2 : 1.125)) >>> 0
		);
		if (prevCapacity != 0) newCapacity = Math.max(newCapacity, 256);
		var oldContents = node.contents;
		node.contents = new Uint8Array(newCapacity);
		if (node.usedBytes > 0) node.contents.set(oldContents.subarray(0, node.usedBytes), 0);
	},
	resizeFileStorage(node, newSize) {
		if (node.usedBytes == newSize) return;
		if (newSize == 0) {
			node.contents = null;
			node.usedBytes = 0;
		} else {
			var oldContents = node.contents;
			node.contents = new Uint8Array(newSize);
			if (oldContents) {
				node.contents.set(oldContents.subarray(0, Math.min(newSize, node.usedBytes)));
			}
			node.usedBytes = newSize;
		}
	},
	node_ops: {
		getattr(node) {
			var attr = {};
			attr.dev = FS.isChrdev(node.mode) ? node.id : 1;
			attr.ino = node.id;
			attr.mode = node.mode;
			attr.nlink = 1;
			attr.uid = 0;
			attr.gid = 0;
			attr.rdev = node.rdev;
			if (FS.isDir(node.mode)) {
				attr.size = 4096;
			} else if (FS.isFile(node.mode)) {
				attr.size = node.usedBytes;
			} else if (FS.isLink(node.mode)) {
				attr.size = node.link.length;
			} else {
				attr.size = 0;
			}
			attr.atime = new Date(node.timestamp);
			attr.mtime = new Date(node.timestamp);
			attr.ctime = new Date(node.timestamp);
			attr.blksize = 4096;
			attr.blocks = Math.ceil(attr.size / attr.blksize);
			return attr;
		},
		setattr(node, attr) {
			if (attr.mode !== undefined) {
				node.mode = attr.mode;
			}
			if (attr.timestamp !== undefined) {
				node.timestamp = attr.timestamp;
			}
			if (attr.size !== undefined) {
				MEMFS.resizeFileStorage(node, attr.size);
			}
		},
		lookup(parent, name) {
			throw FS.genericErrors[44];
		},
		mknod(parent, name, mode, dev) {
			return MEMFS.createNode(parent, name, mode, dev);
		},
		rename(old_node, new_dir, new_name) {
			if (FS.isDir(old_node.mode)) {
				var new_node;
				try {
					new_node = FS.lookupNode(new_dir, new_name);
				} catch (e) {}
				if (new_node) {
					for (var i in new_node.contents) {
						throw new FS.ErrnoError(55);
					}
				}
			}
			delete old_node.parent.contents[old_node.name];
			old_node.parent.timestamp = Date.now();
			old_node.name = new_name;
			new_dir.contents[new_name] = old_node;
			new_dir.timestamp = old_node.parent.timestamp;
			old_node.parent = new_dir;
		},
		unlink(parent, name) {
			delete parent.contents[name];
			parent.timestamp = Date.now();
		},
		rmdir(parent, name) {
			var node = FS.lookupNode(parent, name);
			for (var i in node.contents) {
				throw new FS.ErrnoError(55);
			}
			delete parent.contents[name];
			parent.timestamp = Date.now();
		},
		readdir(node) {
			var entries = ['.', '..'];
			for (var key of Object.keys(node.contents)) {
				entries.push(key);
			}
			return entries;
		},
		symlink(parent, newname, oldpath) {
			var node = MEMFS.createNode(parent, newname, 511 | 40960, 0);
			node.link = oldpath;
			return node;
		},
		readlink(node) {
			if (!FS.isLink(node.mode)) {
				throw new FS.ErrnoError(28);
			}
			return node.link;
		}
	},
	stream_ops: {
		read(stream, buffer, offset, length, position) {
			var contents = stream.node.contents;
			if (position >= stream.node.usedBytes) return 0;
			var size = Math.min(stream.node.usedBytes - position, length);
			assert(size >= 0);
			if (size > 8 && contents.subarray) {
				buffer.set(contents.subarray(position, position + size), offset);
			} else {
				for (var i = 0; i < size; i++) buffer[offset + i] = contents[position + i];
			}
			return size;
		},
		write(stream, buffer, offset, length, position, canOwn) {
			assert(!(buffer instanceof ArrayBuffer));
			if (!length) return 0;
			var node = stream.node;
			node.timestamp = Date.now();
			if (buffer.subarray && (!node.contents || node.contents.subarray)) {
				if (canOwn) {
					assert(position === 0, 'canOwn must imply no weird position inside the file');
					node.contents = buffer.subarray(offset, offset + length);
					node.usedBytes = length;
					return length;
				} else if (node.usedBytes === 0 && position === 0) {
					node.contents = buffer.slice(offset, offset + length);
					node.usedBytes = length;
					return length;
				} else if (position + length <= node.usedBytes) {
					node.contents.set(buffer.subarray(offset, offset + length), position);
					return length;
				}
			}
			MEMFS.expandFileStorage(node, position + length);
			if (node.contents.subarray && buffer.subarray) {
				node.contents.set(buffer.subarray(offset, offset + length), position);
			} else {
				for (var i = 0; i < length; i++) {
					node.contents[position + i] = buffer[offset + i];
				}
			}
			node.usedBytes = Math.max(node.usedBytes, position + length);
			return length;
		},
		llseek(stream, offset, whence) {
			var position = offset;
			if (whence === 1) {
				position += stream.position;
			} else if (whence === 2) {
				if (FS.isFile(stream.node.mode)) {
					position += stream.node.usedBytes;
				}
			}
			if (position < 0) {
				throw new FS.ErrnoError(28);
			}
			return position;
		},
		allocate(stream, offset, length) {
			MEMFS.expandFileStorage(stream.node, offset + length);
			stream.node.usedBytes = Math.max(stream.node.usedBytes, offset + length);
		},
		mmap(stream, length, position, prot, flags) {
			if (!FS.isFile(stream.node.mode)) {
				throw new FS.ErrnoError(43);
			}
			var ptr;
			var allocated;
			var contents = stream.node.contents;
			if (!(flags & 2) && contents.buffer === HEAP8.buffer) {
				allocated = false;
				ptr = contents.byteOffset;
			} else {
				if (position > 0 || position + length < contents.length) {
					if (contents.subarray) {
						contents = contents.subarray(position, position + length);
					} else {
						contents = Array.prototype.slice.call(contents, position, position + length);
					}
				}
				allocated = true;
				ptr = mmapAlloc(length);
				if (!ptr) {
					throw new FS.ErrnoError(48);
				}
				HEAP8.set(contents, ptr);
			}
			return { ptr: ptr, allocated: allocated };
		},
		msync(stream, buffer, offset, length, mmapFlags) {
			MEMFS.stream_ops.write(stream, buffer, 0, length, offset, false);
			return 0;
		}
	}
};
var asyncLoad = (url, onload, onerror, noRunDep) => {
	var dep = !noRunDep ? getUniqueRunDependency(`al ${url}`) : '';
	readAsync(
		url,
		(arrayBuffer) => {
			assert(arrayBuffer, `Loading data file "${url}" failed (no arrayBuffer).`);
			onload(new Uint8Array(arrayBuffer));
			if (dep) removeRunDependency(dep);
		},
		(event) => {
			if (onerror) {
				onerror();
			} else {
				throw `Loading data file "${url}" failed.`;
			}
		}
	);
	if (dep) addRunDependency(dep);
};
var FS_createDataFile = (parent, name, fileData, canRead, canWrite, canOwn) => {
	FS.createDataFile(parent, name, fileData, canRead, canWrite, canOwn);
};
var preloadPlugins = Module['preloadPlugins'] || [];
var FS_handledByPreloadPlugin = (byteArray, fullname, finish, onerror) => {
	if (typeof Browser != 'undefined') Browser.init();
	var handled = false;
	preloadPlugins.forEach((plugin) => {
		if (handled) return;
		if (plugin['canHandle'](fullname)) {
			plugin['handle'](byteArray, fullname, finish, onerror);
			handled = true;
		}
	});
	return handled;
};
var FS_createPreloadedFile = (
	parent,
	name,
	url,
	canRead,
	canWrite,
	onload,
	onerror,
	dontCreateFile,
	canOwn,
	preFinish
) => {
	var fullname = name ? PATH_FS.resolve(PATH.join2(parent, name)) : parent;
	var dep = getUniqueRunDependency(`cp ${fullname}`);
	function processData(byteArray) {
		function finish(byteArray) {
			preFinish?.();
			if (!dontCreateFile) {
				FS_createDataFile(parent, name, byteArray, canRead, canWrite, canOwn);
			}
			onload?.();
			removeRunDependency(dep);
		}
		if (
			FS_handledByPreloadPlugin(byteArray, fullname, finish, () => {
				onerror?.();
				removeRunDependency(dep);
			})
		) {
			return;
		}
		finish(byteArray);
	}
	addRunDependency(dep);
	if (typeof url == 'string') {
		asyncLoad(url, (byteArray) => processData(byteArray), onerror);
	} else {
		processData(url);
	}
};
var FS_modeStringToFlags = (str) => {
	var flagModes = {
		r: 0,
		'r+': 2,
		w: 512 | 64 | 1,
		'w+': 512 | 64 | 2,
		a: 1024 | 64 | 1,
		'a+': 1024 | 64 | 2
	};
	var flags = flagModes[str];
	if (typeof flags == 'undefined') {
		throw new Error(`Unknown file open mode: ${str}`);
	}
	return flags;
};
var FS_getMode = (canRead, canWrite) => {
	var mode = 0;
	if (canRead) mode |= 292 | 73;
	if (canWrite) mode |= 146;
	return mode;
};
var ERRNO_MESSAGES = {
	0: 'Success',
	1: 'Arg list too long',
	2: 'Permission denied',
	3: 'Address already in use',
	4: 'Address not available',
	5: 'Address family not supported by protocol family',
	6: 'No more processes',
	7: 'Socket already connected',
	8: 'Bad file number',
	9: 'Trying to read unreadable message',
	10: 'Mount device busy',
	11: 'Operation canceled',
	12: 'No children',
	13: 'Connection aborted',
	14: 'Connection refused',
	15: 'Connection reset by peer',
	16: 'File locking deadlock error',
	17: 'Destination address required',
	18: 'Math arg out of domain of func',
	19: 'Quota exceeded',
	20: 'File exists',
	21: 'Bad address',
	22: 'File too large',
	23: 'Host is unreachable',
	24: 'Identifier removed',
	25: 'Illegal byte sequence',
	26: 'Connection already in progress',
	27: 'Interrupted system call',
	28: 'Invalid argument',
	29: 'I/O error',
	30: 'Socket is already connected',
	31: 'Is a directory',
	32: 'Too many symbolic links',
	33: 'Too many open files',
	34: 'Too many links',
	35: 'Message too long',
	36: 'Multihop attempted',
	37: 'File or path name too long',
	38: 'Network interface is not configured',
	39: 'Connection reset by network',
	40: 'Network is unreachable',
	41: 'Too many open files in system',
	42: 'No buffer space available',
	43: 'No such device',
	44: 'No such file or directory',
	45: 'Exec format error',
	46: 'No record locks available',
	47: 'The link has been severed',
	48: 'Not enough core',
	49: 'No message of desired type',
	50: 'Protocol not available',
	51: 'No space left on device',
	52: 'Function not implemented',
	53: 'Socket is not connected',
	54: 'Not a directory',
	55: 'Directory not empty',
	56: 'State not recoverable',
	57: 'Socket operation on non-socket',
	59: 'Not a typewriter',
	60: 'No such device or address',
	61: 'Value too large for defined data type',
	62: 'Previous owner died',
	63: 'Not super-user',
	64: 'Broken pipe',
	65: 'Protocol error',
	66: 'Unknown protocol',
	67: 'Protocol wrong type for socket',
	68: 'Math result not representable',
	69: 'Read only file system',
	70: 'Illegal seek',
	71: 'No such process',
	72: 'Stale file handle',
	73: 'Connection timed out',
	74: 'Text file busy',
	75: 'Cross-device link',
	100: 'Device not a stream',
	101: 'Bad font file fmt',
	102: 'Invalid slot',
	103: 'Invalid request code',
	104: 'No anode',
	105: 'Block device required',
	106: 'Channel number out of range',
	107: 'Level 3 halted',
	108: 'Level 3 reset',
	109: 'Link number out of range',
	110: 'Protocol driver not attached',
	111: 'No CSI structure available',
	112: 'Level 2 halted',
	113: 'Invalid exchange',
	114: 'Invalid request descriptor',
	115: 'Exchange full',
	116: 'No data (for no delay io)',
	117: 'Timer expired',
	118: 'Out of streams resources',
	119: 'Machine is not on the network',
	120: 'Package not installed',
	121: 'The object is remote',
	122: 'Advertise error',
	123: 'Srmount error',
	124: 'Communication error on send',
	125: 'Cross mount point (not really error)',
	126: 'Given log. name not unique',
	127: 'f.d. invalid for this operation',
	128: 'Remote address changed',
	129: 'Can   access a needed shared lib',
	130: 'Accessing a corrupted shared lib',
	131: '.lib section in a.out corrupted',
	132: 'Attempting to link in too many libs',
	133: 'Attempting to exec a shared library',
	135: 'Streams pipe error',
	136: 'Too many users',
	137: 'Socket type not supported',
	138: 'Not supported',
	139: 'Protocol family not supported',
	140: "Can't send after socket shutdown",
	141: 'Too many references',
	142: 'Host is down',
	148: 'No medium (in tape drive)',
	156: 'Level 2 not synchronized'
};
var ERRNO_CODES = {
	EPERM: 63,
	ENOENT: 44,
	ESRCH: 71,
	EINTR: 27,
	EIO: 29,
	ENXIO: 60,
	E2BIG: 1,
	ENOEXEC: 45,
	EBADF: 8,
	ECHILD: 12,
	EAGAIN: 6,
	EWOULDBLOCK: 6,
	ENOMEM: 48,
	EACCES: 2,
	EFAULT: 21,
	ENOTBLK: 105,
	EBUSY: 10,
	EEXIST: 20,
	EXDEV: 75,
	ENODEV: 43,
	ENOTDIR: 54,
	EISDIR: 31,
	EINVAL: 28,
	ENFILE: 41,
	EMFILE: 33,
	ENOTTY: 59,
	ETXTBSY: 74,
	EFBIG: 22,
	ENOSPC: 51,
	ESPIPE: 70,
	EROFS: 69,
	EMLINK: 34,
	EPIPE: 64,
	EDOM: 18,
	ERANGE: 68,
	ENOMSG: 49,
	EIDRM: 24,
	ECHRNG: 106,
	EL2NSYNC: 156,
	EL3HLT: 107,
	EL3RST: 108,
	ELNRNG: 109,
	EUNATCH: 110,
	ENOCSI: 111,
	EL2HLT: 112,
	EDEADLK: 16,
	ENOLCK: 46,
	EBADE: 113,
	EBADR: 114,
	EXFULL: 115,
	ENOANO: 104,
	EBADRQC: 103,
	EBADSLT: 102,
	EDEADLOCK: 16,
	EBFONT: 101,
	ENOSTR: 100,
	ENODATA: 116,
	ETIME: 117,
	ENOSR: 118,
	ENONET: 119,
	ENOPKG: 120,
	EREMOTE: 121,
	ENOLINK: 47,
	EADV: 122,
	ESRMNT: 123,
	ECOMM: 124,
	EPROTO: 65,
	EMULTIHOP: 36,
	EDOTDOT: 125,
	EBADMSG: 9,
	ENOTUNIQ: 126,
	EBADFD: 127,
	EREMCHG: 128,
	ELIBACC: 129,
	ELIBBAD: 130,
	ELIBSCN: 131,
	ELIBMAX: 132,
	ELIBEXEC: 133,
	ENOSYS: 52,
	ENOTEMPTY: 55,
	ENAMETOOLONG: 37,
	ELOOP: 32,
	EOPNOTSUPP: 138,
	EPFNOSUPPORT: 139,
	ECONNRESET: 15,
	ENOBUFS: 42,
	EAFNOSUPPORT: 5,
	EPROTOTYPE: 67,
	ENOTSOCK: 57,
	ENOPROTOOPT: 50,
	ESHUTDOWN: 140,
	ECONNREFUSED: 14,
	EADDRINUSE: 3,
	ECONNABORTED: 13,
	ENETUNREACH: 40,
	ENETDOWN: 38,
	ETIMEDOUT: 73,
	EHOSTDOWN: 142,
	EHOSTUNREACH: 23,
	EINPROGRESS: 26,
	EALREADY: 7,
	EDESTADDRREQ: 17,
	EMSGSIZE: 35,
	EPROTONOSUPPORT: 66,
	ESOCKTNOSUPPORT: 137,
	EADDRNOTAVAIL: 4,
	ENETRESET: 39,
	EISCONN: 30,
	ENOTCONN: 53,
	ETOOMANYREFS: 141,
	EUSERS: 136,
	EDQUOT: 19,
	ESTALE: 72,
	ENOTSUP: 138,
	ENOMEDIUM: 148,
	EILSEQ: 25,
	EOVERFLOW: 61,
	ECANCELED: 11,
	ENOTRECOVERABLE: 56,
	EOWNERDEAD: 62,
	ESTRPIPE: 135
};
var demangle = (func) => {
	warnOnce('warning: build with -sDEMANGLE_SUPPORT to link in libcxxabi demangling');
	return func;
};
var demangleAll = (text) => {
	var regex = /\b_Z[\w\d_]+/g;
	return text.replace(regex, function (x) {
		var y = demangle(x);
		return x === y ? x : y + ' [' + x + ']';
	});
};
var FS = {
	root: null,
	mounts: [],
	devices: {},
	streams: [],
	nextInode: 1,
	nameTable: null,
	currentPath: '/',
	initialized: false,
	ignorePermissions: true,
	ErrnoError: null,
	genericErrors: {},
	filesystems: null,
	syncFSRequests: 0,
	lookupPath(path, opts = {}) {
		path = PATH_FS.resolve(path);
		if (!path) return { path: '', node: null };
		var defaults = { follow_mount: true, recurse_count: 0 };
		opts = Object.assign(defaults, opts);
		if (opts.recurse_count > 8) {
			throw new FS.ErrnoError(32);
		}
		var parts = path.split('/').filter((p) => !!p);
		var current = FS.root;
		var current_path = '/';
		for (var i = 0; i < parts.length; i++) {
			var islast = i === parts.length - 1;
			if (islast && opts.parent) {
				break;
			}
			current = FS.lookupNode(current, parts[i]);
			current_path = PATH.join2(current_path, parts[i]);
			if (FS.isMountpoint(current)) {
				if (!islast || (islast && opts.follow_mount)) {
					current = current.mounted.root;
				}
			}
			if (!islast || opts.follow) {
				var count = 0;
				while (FS.isLink(current.mode)) {
					var link = FS.readlink(current_path);
					current_path = PATH_FS.resolve(PATH.dirname(current_path), link);
					var lookup = FS.lookupPath(current_path, { recurse_count: opts.recurse_count + 1 });
					current = lookup.node;
					if (count++ > 40) {
						throw new FS.ErrnoError(32);
					}
				}
			}
		}
		return { path: current_path, node: current };
	},
	getPath(node) {
		var path;
		while (true) {
			if (FS.isRoot(node)) {
				var mount = node.mount.mountpoint;
				if (!path) return mount;
				return mount[mount.length - 1] !== '/' ? `${mount}/${path}` : mount + path;
			}
			path = path ? `${node.name}/${path}` : node.name;
			node = node.parent;
		}
	},
	hashName(parentid, name) {
		var hash = 0;
		for (var i = 0; i < name.length; i++) {
			hash = ((hash << 5) - hash + name.charCodeAt(i)) | 0;
		}
		return ((parentid + hash) >>> 0) % FS.nameTable.length;
	},
	hashAddNode(node) {
		var hash = FS.hashName(node.parent.id, node.name);
		node.name_next = FS.nameTable[hash];
		FS.nameTable[hash] = node;
	},
	hashRemoveNode(node) {
		var hash = FS.hashName(node.parent.id, node.name);
		if (FS.nameTable[hash] === node) {
			FS.nameTable[hash] = node.name_next;
		} else {
			var current = FS.nameTable[hash];
			while (current) {
				if (current.name_next === node) {
					current.name_next = node.name_next;
					break;
				}
				current = current.name_next;
			}
		}
	},
	lookupNode(parent, name) {
		var errCode = FS.mayLookup(parent);
		if (errCode) {
			throw new FS.ErrnoError(errCode, parent);
		}
		var hash = FS.hashName(parent.id, name);
		for (var node = FS.nameTable[hash]; node; node = node.name_next) {
			var nodeName = node.name;
			if (node.parent.id === parent.id && nodeName === name) {
				return node;
			}
		}
		return FS.lookup(parent, name);
	},
	createNode(parent, name, mode, rdev) {
		assert(typeof parent == 'object');
		var node = new FS.FSNode(parent, name, mode, rdev);
		FS.hashAddNode(node);
		return node;
	},
	destroyNode(node) {
		FS.hashRemoveNode(node);
	},
	isRoot(node) {
		return node === node.parent;
	},
	isMountpoint(node) {
		return !!node.mounted;
	},
	isFile(mode) {
		return (mode & 61440) === 32768;
	},
	isDir(mode) {
		return (mode & 61440) === 16384;
	},
	isLink(mode) {
		return (mode & 61440) === 40960;
	},
	isChrdev(mode) {
		return (mode & 61440) === 8192;
	},
	isBlkdev(mode) {
		return (mode & 61440) === 24576;
	},
	isFIFO(mode) {
		return (mode & 61440) === 4096;
	},
	isSocket(mode) {
		return (mode & 49152) === 49152;
	},
	flagsToPermissionString(flag) {
		var perms = ['r', 'w', 'rw'][flag & 3];
		if (flag & 512) {
			perms += 'w';
		}
		return perms;
	},
	nodePermissions(node, perms) {
		if (FS.ignorePermissions) {
			return 0;
		}
		if (perms.includes('r') && !(node.mode & 292)) {
			return 2;
		} else if (perms.includes('w') && !(node.mode & 146)) {
			return 2;
		} else if (perms.includes('x') && !(node.mode & 73)) {
			return 2;
		}
		return 0;
	},
	mayLookup(dir) {
		var errCode = FS.nodePermissions(dir, 'x');
		if (errCode) return errCode;
		if (!dir.node_ops.lookup) return 2;
		return 0;
	},
	mayCreate(dir, name) {
		try {
			var node = FS.lookupNode(dir, name);
			return 20;
		} catch (e) {}
		return FS.nodePermissions(dir, 'wx');
	},
	mayDelete(dir, name, isdir) {
		var node;
		try {
			node = FS.lookupNode(dir, name);
		} catch (e) {
			return e.errno;
		}
		var errCode = FS.nodePermissions(dir, 'wx');
		if (errCode) {
			return errCode;
		}
		if (isdir) {
			if (!FS.isDir(node.mode)) {
				return 54;
			}
			if (FS.isRoot(node) || FS.getPath(node) === FS.cwd()) {
				return 10;
			}
		} else {
			if (FS.isDir(node.mode)) {
				return 31;
			}
		}
		return 0;
	},
	mayOpen(node, flags) {
		if (!node) {
			return 44;
		}
		if (FS.isLink(node.mode)) {
			return 32;
		} else if (FS.isDir(node.mode)) {
			if (FS.flagsToPermissionString(flags) !== 'r' || flags & 512) {
				return 31;
			}
		}
		return FS.nodePermissions(node, FS.flagsToPermissionString(flags));
	},
	MAX_OPEN_FDS: 4096,
	nextfd() {
		for (var fd = 0; fd <= FS.MAX_OPEN_FDS; fd++) {
			if (!FS.streams[fd]) {
				return fd;
			}
		}
		throw new FS.ErrnoError(33);
	},
	getStreamChecked(fd) {
		var stream = FS.getStream(fd);
		if (!stream) {
			throw new FS.ErrnoError(8);
		}
		return stream;
	},
	getStream: (fd) => FS.streams[fd],
	createStream(stream, fd = -1) {
		if (!FS.FSStream) {
			FS.FSStream = function () {
				this.shared = {};
			};
			FS.FSStream.prototype = {};
			Object.defineProperties(FS.FSStream.prototype, {
				object: {
					get() {
						return this.node;
					},
					set(val) {
						this.node = val;
					}
				},
				isRead: {
					get() {
						return (this.flags & 2097155) !== 1;
					}
				},
				isWrite: {
					get() {
						return (this.flags & 2097155) !== 0;
					}
				},
				isAppend: {
					get() {
						return this.flags & 1024;
					}
				},
				flags: {
					get() {
						return this.shared.flags;
					},
					set(val) {
						this.shared.flags = val;
					}
				},
				position: {
					get() {
						return this.shared.position;
					},
					set(val) {
						this.shared.position = val;
					}
				}
			});
		}
		stream = Object.assign(new FS.FSStream(), stream);
		if (fd == -1) {
			fd = FS.nextfd();
		}
		stream.fd = fd;
		FS.streams[fd] = stream;
		return stream;
	},
	closeStream(fd) {
		FS.streams[fd] = null;
	},
	chrdev_stream_ops: {
		open(stream) {
			var device = FS.getDevice(stream.node.rdev);
			stream.stream_ops = device.stream_ops;
			stream.stream_ops.open?.(stream);
		},
		llseek() {
			throw new FS.ErrnoError(70);
		}
	},
	major: (dev) => dev >> 8,
	minor: (dev) => dev & 255,
	makedev: (ma, mi) => (ma << 8) | mi,
	registerDevice(dev, ops) {
		FS.devices[dev] = { stream_ops: ops };
	},
	getDevice: (dev) => FS.devices[dev],
	getMounts(mount) {
		var mounts = [];
		var check = [mount];
		while (check.length) {
			var m = check.pop();
			mounts.push(m);
			check.push.apply(check, m.mounts);
		}
		return mounts;
	},
	syncfs(populate, callback) {
		if (typeof populate == 'function') {
			callback = populate;
			populate = false;
		}
		FS.syncFSRequests++;
		if (FS.syncFSRequests > 1) {
			err(
				`warning: ${FS.syncFSRequests} FS.syncfs operations in flight at once, probably just doing extra work`
			);
		}
		var mounts = FS.getMounts(FS.root.mount);
		var completed = 0;
		function doCallback(errCode) {
			assert(FS.syncFSRequests > 0);
			FS.syncFSRequests--;
			return callback(errCode);
		}
		function done(errCode) {
			if (errCode) {
				if (!done.errored) {
					done.errored = true;
					return doCallback(errCode);
				}
				return;
			}
			if (++completed >= mounts.length) {
				doCallback(null);
			}
		}
		mounts.forEach((mount) => {
			if (!mount.type.syncfs) {
				return done(null);
			}
			mount.type.syncfs(mount, populate, done);
		});
	},
	mount(type, opts, mountpoint) {
		if (typeof type == 'string') {
			throw type;
		}
		var root = mountpoint === '/';
		var pseudo = !mountpoint;
		var node;
		if (root && FS.root) {
			throw new FS.ErrnoError(10);
		} else if (!root && !pseudo) {
			var lookup = FS.lookupPath(mountpoint, { follow_mount: false });
			mountpoint = lookup.path;
			node = lookup.node;
			if (FS.isMountpoint(node)) {
				throw new FS.ErrnoError(10);
			}
			if (!FS.isDir(node.mode)) {
				throw new FS.ErrnoError(54);
			}
		}
		var mount = { type: type, opts: opts, mountpoint: mountpoint, mounts: [] };
		var mountRoot = type.mount(mount);
		mountRoot.mount = mount;
		mount.root = mountRoot;
		if (root) {
			FS.root = mountRoot;
		} else if (node) {
			node.mounted = mount;
			if (node.mount) {
				node.mount.mounts.push(mount);
			}
		}
		return mountRoot;
	},
	unmount(mountpoint) {
		var lookup = FS.lookupPath(mountpoint, { follow_mount: false });
		if (!FS.isMountpoint(lookup.node)) {
			throw new FS.ErrnoError(28);
		}
		var node = lookup.node;
		var mount = node.mounted;
		var mounts = FS.getMounts(mount);
		Object.keys(FS.nameTable).forEach((hash) => {
			var current = FS.nameTable[hash];
			while (current) {
				var next = current.name_next;
				if (mounts.includes(current.mount)) {
					FS.destroyNode(current);
				}
				current = next;
			}
		});
		node.mounted = null;
		var idx = node.mount.mounts.indexOf(mount);
		assert(idx !== -1);
		node.mount.mounts.splice(idx, 1);
	},
	lookup(parent, name) {
		return parent.node_ops.lookup(parent, name);
	},
	mknod(path, mode, dev) {
		var lookup = FS.lookupPath(path, { parent: true });
		var parent = lookup.node;
		var name = PATH.basename(path);
		if (!name || name === '.' || name === '..') {
			throw new FS.ErrnoError(28);
		}
		var errCode = FS.mayCreate(parent, name);
		if (errCode) {
			throw new FS.ErrnoError(errCode);
		}
		if (!parent.node_ops.mknod) {
			throw new FS.ErrnoError(63);
		}
		return parent.node_ops.mknod(parent, name, mode, dev);
	},
	create(path, mode) {
		mode = mode !== undefined ? mode : 438;
		mode &= 4095;
		mode |= 32768;
		return FS.mknod(path, mode, 0);
	},
	mkdir(path, mode) {
		mode = mode !== undefined ? mode : 511;
		mode &= 511 | 512;
		mode |= 16384;
		return FS.mknod(path, mode, 0);
	},
	mkdirTree(path, mode) {
		var dirs = path.split('/');
		var d = '';
		for (var i = 0; i < dirs.length; ++i) {
			if (!dirs[i]) continue;
			d += '/' + dirs[i];
			try {
				FS.mkdir(d, mode);
			} catch (e) {
				if (e.errno != 20) throw e;
			}
		}
	},
	mkdev(path, mode, dev) {
		if (typeof dev == 'undefined') {
			dev = mode;
			mode = 438;
		}
		mode |= 8192;
		return FS.mknod(path, mode, dev);
	},
	symlink(oldpath, newpath) {
		if (!PATH_FS.resolve(oldpath)) {
			throw new FS.ErrnoError(44);
		}
		var lookup = FS.lookupPath(newpath, { parent: true });
		var parent = lookup.node;
		if (!parent) {
			throw new FS.ErrnoError(44);
		}
		var newname = PATH.basename(newpath);
		var errCode = FS.mayCreate(parent, newname);
		if (errCode) {
			throw new FS.ErrnoError(errCode);
		}
		if (!parent.node_ops.symlink) {
			throw new FS.ErrnoError(63);
		}
		return parent.node_ops.symlink(parent, newname, oldpath);
	},
	rename(old_path, new_path) {
		var old_dirname = PATH.dirname(old_path);
		var new_dirname = PATH.dirname(new_path);
		var old_name = PATH.basename(old_path);
		var new_name = PATH.basename(new_path);
		var lookup, old_dir, new_dir;
		lookup = FS.lookupPath(old_path, { parent: true });
		old_dir = lookup.node;
		lookup = FS.lookupPath(new_path, { parent: true });
		new_dir = lookup.node;
		if (!old_dir || !new_dir) throw new FS.ErrnoError(44);
		if (old_dir.mount !== new_dir.mount) {
			throw new FS.ErrnoError(75);
		}
		var old_node = FS.lookupNode(old_dir, old_name);
		var relative = PATH_FS.relative(old_path, new_dirname);
		if (relative.charAt(0) !== '.') {
			throw new FS.ErrnoError(28);
		}
		relative = PATH_FS.relative(new_path, old_dirname);
		if (relative.charAt(0) !== '.') {
			throw new FS.ErrnoError(55);
		}
		var new_node;
		try {
			new_node = FS.lookupNode(new_dir, new_name);
		} catch (e) {}
		if (old_node === new_node) {
			return;
		}
		var isdir = FS.isDir(old_node.mode);
		var errCode = FS.mayDelete(old_dir, old_name, isdir);
		if (errCode) {
			throw new FS.ErrnoError(errCode);
		}
		errCode = new_node ? FS.mayDelete(new_dir, new_name, isdir) : FS.mayCreate(new_dir, new_name);
		if (errCode) {
			throw new FS.ErrnoError(errCode);
		}
		if (!old_dir.node_ops.rename) {
			throw new FS.ErrnoError(63);
		}
		if (FS.isMountpoint(old_node) || (new_node && FS.isMountpoint(new_node))) {
			throw new FS.ErrnoError(10);
		}
		if (new_dir !== old_dir) {
			errCode = FS.nodePermissions(old_dir, 'w');
			if (errCode) {
				throw new FS.ErrnoError(errCode);
			}
		}
		FS.hashRemoveNode(old_node);
		try {
			old_dir.node_ops.rename(old_node, new_dir, new_name);
		} catch (e) {
			throw e;
		} finally {
			FS.hashAddNode(old_node);
		}
	},
	rmdir(path) {
		var lookup = FS.lookupPath(path, { parent: true });
		var parent = lookup.node;
		var name = PATH.basename(path);
		var node = FS.lookupNode(parent, name);
		var errCode = FS.mayDelete(parent, name, true);
		if (errCode) {
			throw new FS.ErrnoError(errCode);
		}
		if (!parent.node_ops.rmdir) {
			throw new FS.ErrnoError(63);
		}
		if (FS.isMountpoint(node)) {
			throw new FS.ErrnoError(10);
		}
		parent.node_ops.rmdir(parent, name);
		FS.destroyNode(node);
	},
	readdir(path) {
		var lookup = FS.lookupPath(path, { follow: true });
		var node = lookup.node;
		if (!node.node_ops.readdir) {
			throw new FS.ErrnoError(54);
		}
		return node.node_ops.readdir(node);
	},
	unlink(path) {
		var lookup = FS.lookupPath(path, { parent: true });
		var parent = lookup.node;
		if (!parent) {
			throw new FS.ErrnoError(44);
		}
		var name = PATH.basename(path);
		var node = FS.lookupNode(parent, name);
		var errCode = FS.mayDelete(parent, name, false);
		if (errCode) {
			throw new FS.ErrnoError(errCode);
		}
		if (!parent.node_ops.unlink) {
			throw new FS.ErrnoError(63);
		}
		if (FS.isMountpoint(node)) {
			throw new FS.ErrnoError(10);
		}
		parent.node_ops.unlink(parent, name);
		FS.destroyNode(node);
	},
	readlink(path) {
		var lookup = FS.lookupPath(path);
		var link = lookup.node;
		if (!link) {
			throw new FS.ErrnoError(44);
		}
		if (!link.node_ops.readlink) {
			throw new FS.ErrnoError(28);
		}
		return PATH_FS.resolve(FS.getPath(link.parent), link.node_ops.readlink(link));
	},
	stat(path, dontFollow) {
		var lookup = FS.lookupPath(path, { follow: !dontFollow });
		var node = lookup.node;
		if (!node) {
			throw new FS.ErrnoError(44);
		}
		if (!node.node_ops.getattr) {
			throw new FS.ErrnoError(63);
		}
		return node.node_ops.getattr(node);
	},
	lstat(path) {
		return FS.stat(path, true);
	},
	chmod(path, mode, dontFollow) {
		var node;
		if (typeof path == 'string') {
			var lookup = FS.lookupPath(path, { follow: !dontFollow });
			node = lookup.node;
		} else {
			node = path;
		}
		if (!node.node_ops.setattr) {
			throw new FS.ErrnoError(63);
		}
		node.node_ops.setattr(node, {
			mode: (mode & 4095) | (node.mode & ~4095),
			timestamp: Date.now()
		});
	},
	lchmod(path, mode) {
		FS.chmod(path, mode, true);
	},
	fchmod(fd, mode) {
		var stream = FS.getStreamChecked(fd);
		FS.chmod(stream.node, mode);
	},
	chown(path, uid, gid, dontFollow) {
		var node;
		if (typeof path == 'string') {
			var lookup = FS.lookupPath(path, { follow: !dontFollow });
			node = lookup.node;
		} else {
			node = path;
		}
		if (!node.node_ops.setattr) {
			throw new FS.ErrnoError(63);
		}
		node.node_ops.setattr(node, { timestamp: Date.now() });
	},
	lchown(path, uid, gid) {
		FS.chown(path, uid, gid, true);
	},
	fchown(fd, uid, gid) {
		var stream = FS.getStreamChecked(fd);
		FS.chown(stream.node, uid, gid);
	},
	truncate(path, len) {
		if (len < 0) {
			throw new FS.ErrnoError(28);
		}
		var node;
		if (typeof path == 'string') {
			var lookup = FS.lookupPath(path, { follow: true });
			node = lookup.node;
		} else {
			node = path;
		}
		if (!node.node_ops.setattr) {
			throw new FS.ErrnoError(63);
		}
		if (FS.isDir(node.mode)) {
			throw new FS.ErrnoError(31);
		}
		if (!FS.isFile(node.mode)) {
			throw new FS.ErrnoError(28);
		}
		var errCode = FS.nodePermissions(node, 'w');
		if (errCode) {
			throw new FS.ErrnoError(errCode);
		}
		node.node_ops.setattr(node, { size: len, timestamp: Date.now() });
	},
	ftruncate(fd, len) {
		var stream = FS.getStreamChecked(fd);
		if ((stream.flags & 2097155) === 0) {
			throw new FS.ErrnoError(28);
		}
		FS.truncate(stream.node, len);
	},
	utime(path, atime, mtime) {
		var lookup = FS.lookupPath(path, { follow: true });
		var node = lookup.node;
		node.node_ops.setattr(node, { timestamp: Math.max(atime, mtime) });
	},
	open(path, flags, mode) {
		if (path === '') {
			throw new FS.ErrnoError(44);
		}
		flags = typeof flags == 'string' ? FS_modeStringToFlags(flags) : flags;
		mode = typeof mode == 'undefined' ? 438 : mode;
		if (flags & 64) {
			mode = (mode & 4095) | 32768;
		} else {
			mode = 0;
		}
		var node;
		if (typeof path == 'object') {
			node = path;
		} else {
			path = PATH.normalize(path);
			try {
				var lookup = FS.lookupPath(path, { follow: !(flags & 131072) });
				node = lookup.node;
			} catch (e) {}
		}
		var created = false;
		if (flags & 64) {
			if (node) {
				if (flags & 128) {
					throw new FS.ErrnoError(20);
				}
			} else {
				node = FS.mknod(path, mode, 0);
				created = true;
			}
		}
		if (!node) {
			throw new FS.ErrnoError(44);
		}
		if (FS.isChrdev(node.mode)) {
			flags &= ~512;
		}
		if (flags & 65536 && !FS.isDir(node.mode)) {
			throw new FS.ErrnoError(54);
		}
		if (!created) {
			var errCode = FS.mayOpen(node, flags);
			if (errCode) {
				throw new FS.ErrnoError(errCode);
			}
		}
		if (flags & 512 && !created) {
			FS.truncate(node, 0);
		}
		flags &= ~(128 | 512 | 131072);
		var stream = FS.createStream({
			node: node,
			path: FS.getPath(node),
			flags: flags,
			seekable: true,
			position: 0,
			stream_ops: node.stream_ops,
			ungotten: [],
			error: false
		});
		if (stream.stream_ops.open) {
			stream.stream_ops.open(stream);
		}
		if (Module['logReadFiles'] && !(flags & 1)) {
			if (!FS.readFiles) FS.readFiles = {};
			if (!(path in FS.readFiles)) {
				FS.readFiles[path] = 1;
			}
		}
		return stream;
	},
	close(stream) {
		if (FS.isClosed(stream)) {
			throw new FS.ErrnoError(8);
		}
		if (stream.getdents) stream.getdents = null;
		try {
			if (stream.stream_ops.close) {
				stream.stream_ops.close(stream);
			}
		} catch (e) {
			throw e;
		} finally {
			FS.closeStream(stream.fd);
		}
		stream.fd = null;
	},
	isClosed(stream) {
		return stream.fd === null;
	},
	llseek(stream, offset, whence) {
		if (FS.isClosed(stream)) {
			throw new FS.ErrnoError(8);
		}
		if (!stream.seekable || !stream.stream_ops.llseek) {
			throw new FS.ErrnoError(70);
		}
		if (whence != 0 && whence != 1 && whence != 2) {
			throw new FS.ErrnoError(28);
		}
		stream.position = stream.stream_ops.llseek(stream, offset, whence);
		stream.ungotten = [];
		return stream.position;
	},
	read(stream, buffer, offset, length, position) {
		assert(offset >= 0);
		if (length < 0 || position < 0) {
			throw new FS.ErrnoError(28);
		}
		if (FS.isClosed(stream)) {
			throw new FS.ErrnoError(8);
		}
		if ((stream.flags & 2097155) === 1) {
			throw new FS.ErrnoError(8);
		}
		if (FS.isDir(stream.node.mode)) {
			throw new FS.ErrnoError(31);
		}
		if (!stream.stream_ops.read) {
			throw new FS.ErrnoError(28);
		}
		var seeking = typeof position != 'undefined';
		if (!seeking) {
			position = stream.position;
		} else if (!stream.seekable) {
			throw new FS.ErrnoError(70);
		}
		var bytesRead = stream.stream_ops.read(stream, buffer, offset, length, position);
		if (!seeking) stream.position += bytesRead;
		return bytesRead;
	},
	write(stream, buffer, offset, length, position, canOwn) {
		assert(offset >= 0);
		if (length < 0 || position < 0) {
			throw new FS.ErrnoError(28);
		}
		if (FS.isClosed(stream)) {
			throw new FS.ErrnoError(8);
		}
		if ((stream.flags & 2097155) === 0) {
			throw new FS.ErrnoError(8);
		}
		if (FS.isDir(stream.node.mode)) {
			throw new FS.ErrnoError(31);
		}
		if (!stream.stream_ops.write) {
			throw new FS.ErrnoError(28);
		}
		if (stream.seekable && stream.flags & 1024) {
			FS.llseek(stream, 0, 2);
		}
		var seeking = typeof position != 'undefined';
		if (!seeking) {
			position = stream.position;
		} else if (!stream.seekable) {
			throw new FS.ErrnoError(70);
		}
		var bytesWritten = stream.stream_ops.write(stream, buffer, offset, length, position, canOwn);
		if (!seeking) stream.position += bytesWritten;
		return bytesWritten;
	},
	allocate(stream, offset, length) {
		if (FS.isClosed(stream)) {
			throw new FS.ErrnoError(8);
		}
		if (offset < 0 || length <= 0) {
			throw new FS.ErrnoError(28);
		}
		if ((stream.flags & 2097155) === 0) {
			throw new FS.ErrnoError(8);
		}
		if (!FS.isFile(stream.node.mode) && !FS.isDir(stream.node.mode)) {
			throw new FS.ErrnoError(43);
		}
		if (!stream.stream_ops.allocate) {
			throw new FS.ErrnoError(138);
		}
		stream.stream_ops.allocate(stream, offset, length);
	},
	mmap(stream, length, position, prot, flags) {
		if ((prot & 2) !== 0 && (flags & 2) === 0 && (stream.flags & 2097155) !== 2) {
			throw new FS.ErrnoError(2);
		}
		if ((stream.flags & 2097155) === 1) {
			throw new FS.ErrnoError(2);
		}
		if (!stream.stream_ops.mmap) {
			throw new FS.ErrnoError(43);
		}
		return stream.stream_ops.mmap(stream, length, position, prot, flags);
	},
	msync(stream, buffer, offset, length, mmapFlags) {
		assert(offset >= 0);
		if (!stream.stream_ops.msync) {
			return 0;
		}
		return stream.stream_ops.msync(stream, buffer, offset, length, mmapFlags);
	},
	munmap: (stream) => 0,
	ioctl(stream, cmd, arg) {
		if (!stream.stream_ops.ioctl) {
			throw new FS.ErrnoError(59);
		}
		return stream.stream_ops.ioctl(stream, cmd, arg);
	},
	readFile(path, opts = {}) {
		opts.flags = opts.flags || 0;
		opts.encoding = opts.encoding || 'binary';
		if (opts.encoding !== 'utf8' && opts.encoding !== 'binary') {
			throw new Error(`Invalid encoding type "${opts.encoding}"`);
		}
		var ret;
		var stream = FS.open(path, opts.flags);
		var stat = FS.stat(path);
		var length = stat.size;
		var buf = new Uint8Array(length);
		FS.read(stream, buf, 0, length, 0);
		if (opts.encoding === 'utf8') {
			ret = UTF8ArrayToString(buf, 0);
		} else if (opts.encoding === 'binary') {
			ret = buf;
		}
		FS.close(stream);
		return ret;
	},
	writeFile(path, data, opts = {}) {
		opts.flags = opts.flags || 577;
		var stream = FS.open(path, opts.flags, opts.mode);
		if (typeof data == 'string') {
			var buf = new Uint8Array(lengthBytesUTF8(data) + 1);
			var actualNumBytes = stringToUTF8Array(data, buf, 0, buf.length);
			FS.write(stream, buf, 0, actualNumBytes, undefined, opts.canOwn);
		} else if (ArrayBuffer.isView(data)) {
			FS.write(stream, data, 0, data.byteLength, undefined, opts.canOwn);
		} else {
			throw new Error('Unsupported data type');
		}
		FS.close(stream);
	},
	cwd: () => FS.currentPath,
	chdir(path) {
		var lookup = FS.lookupPath(path, { follow: true });
		if (lookup.node === null) {
			throw new FS.ErrnoError(44);
		}
		if (!FS.isDir(lookup.node.mode)) {
			throw new FS.ErrnoError(54);
		}
		var errCode = FS.nodePermissions(lookup.node, 'x');
		if (errCode) {
			throw new FS.ErrnoError(errCode);
		}
		FS.currentPath = lookup.path;
	},
	createDefaultDirectories() {
		FS.mkdir('/tmp');
		FS.mkdir('/home');
		FS.mkdir('/home/web_user');
	},
	createDefaultDevices() {
		FS.mkdir('/dev');
		FS.registerDevice(FS.makedev(1, 3), {
			read: () => 0,
			write: (stream, buffer, offset, length, pos) => length
		});
		FS.mkdev('/dev/null', FS.makedev(1, 3));
		TTY.register(FS.makedev(5, 0), TTY.default_tty_ops);
		TTY.register(FS.makedev(6, 0), TTY.default_tty1_ops);
		FS.mkdev('/dev/tty', FS.makedev(5, 0));
		FS.mkdev('/dev/tty1', FS.makedev(6, 0));
		var randomBuffer = new Uint8Array(1024),
			randomLeft = 0;
		var randomByte = () => {
			if (randomLeft === 0) {
				randomLeft = randomFill(randomBuffer).byteLength;
			}
			return randomBuffer[--randomLeft];
		};
		FS.createDevice('/dev', 'random', randomByte);
		FS.createDevice('/dev', 'urandom', randomByte);
		FS.mkdir('/dev/shm');
		FS.mkdir('/dev/shm/tmp');
	},
	createSpecialDirectories() {
		FS.mkdir('/proc');
		var proc_self = FS.mkdir('/proc/self');
		FS.mkdir('/proc/self/fd');
		FS.mount(
			{
				mount() {
					var node = FS.createNode(proc_self, 'fd', 16384 | 511, 73);
					node.node_ops = {
						lookup(parent, name) {
							var fd = +name;
							var stream = FS.getStreamChecked(fd);
							var ret = {
								parent: null,
								mount: { mountpoint: 'fake' },
								node_ops: { readlink: () => stream.path }
							};
							ret.parent = ret;
							return ret;
						}
					};
					return node;
				}
			},
			{},
			'/proc/self/fd'
		);
	},
	createStandardStreams() {
		if (Module['stdin']) {
			FS.createDevice('/dev', 'stdin', Module['stdin']);
		} else {
			FS.symlink('/dev/tty', '/dev/stdin');
		}
		if (Module['stdout']) {
			FS.createDevice('/dev', 'stdout', null, Module['stdout']);
		} else {
			FS.symlink('/dev/tty', '/dev/stdout');
		}
		if (Module['stderr']) {
			FS.createDevice('/dev', 'stderr', null, Module['stderr']);
		} else {
			FS.symlink('/dev/tty1', '/dev/stderr');
		}
		var stdin = FS.open('/dev/stdin', 0);
		var stdout = FS.open('/dev/stdout', 1);
		var stderr = FS.open('/dev/stderr', 1);
		assert(stdin.fd === 0, `invalid handle for stdin (${stdin.fd})`);
		assert(stdout.fd === 1, `invalid handle for stdout (${stdout.fd})`);
		assert(stderr.fd === 2, `invalid handle for stderr (${stderr.fd})`);
	},
	ensureErrnoError() {
		if (FS.ErrnoError) return;
		FS.ErrnoError = function ErrnoError(errno, node) {
			this.name = 'ErrnoError';
			this.node = node;
			this.setErrno = function (errno) {
				this.errno = errno;
				for (var key in ERRNO_CODES) {
					if (ERRNO_CODES[key] === errno) {
						this.code = key;
						break;
					}
				}
			};
			this.setErrno(errno);
			this.message = ERRNO_MESSAGES[errno];
			if (this.stack) {
				Object.defineProperty(this, 'stack', { value: new Error().stack, writable: true });
				this.stack = demangleAll(this.stack);
			}
		};
		FS.ErrnoError.prototype = new Error();
		FS.ErrnoError.prototype.constructor = FS.ErrnoError;
		[44].forEach((code) => {
			FS.genericErrors[code] = new FS.ErrnoError(code);
			FS.genericErrors[code].stack = '<generic error, no stack>';
		});
	},
	staticInit() {
		FS.ensureErrnoError();
		FS.nameTable = new Array(4096);
		FS.mount(MEMFS, {}, '/');
		FS.createDefaultDirectories();
		FS.createDefaultDevices();
		FS.createSpecialDirectories();
		FS.filesystems = { MEMFS: MEMFS };
	},
	init(input, output, error) {
		assert(
			!FS.init.initialized,
			'FS.init was previously called. If you want to initialize later with custom parameters, remove any earlier calls (note that one is automatically added to the generated code)'
		);
		FS.init.initialized = true;
		FS.ensureErrnoError();
		Module['stdin'] = input || Module['stdin'];
		Module['stdout'] = output || Module['stdout'];
		Module['stderr'] = error || Module['stderr'];
		FS.createStandardStreams();
	},
	quit() {
		FS.init.initialized = false;
		_fflush(0);
		for (var i = 0; i < FS.streams.length; i++) {
			var stream = FS.streams[i];
			if (!stream) {
				continue;
			}
			FS.close(stream);
		}
	},
	findObject(path, dontResolveLastLink) {
		var ret = FS.analyzePath(path, dontResolveLastLink);
		if (!ret.exists) {
			return null;
		}
		return ret.object;
	},
	analyzePath(path, dontResolveLastLink) {
		try {
			var lookup = FS.lookupPath(path, { follow: !dontResolveLastLink });
			path = lookup.path;
		} catch (e) {}
		var ret = {
			isRoot: false,
			exists: false,
			error: 0,
			name: null,
			path: null,
			object: null,
			parentExists: false,
			parentPath: null,
			parentObject: null
		};
		try {
			var lookup = FS.lookupPath(path, { parent: true });
			ret.parentExists = true;
			ret.parentPath = lookup.path;
			ret.parentObject = lookup.node;
			ret.name = PATH.basename(path);
			lookup = FS.lookupPath(path, { follow: !dontResolveLastLink });
			ret.exists = true;
			ret.path = lookup.path;
			ret.object = lookup.node;
			ret.name = lookup.node.name;
			ret.isRoot = lookup.path === '/';
		} catch (e) {
			ret.error = e.errno;
		}
		return ret;
	},
	createPath(parent, path, canRead, canWrite) {
		parent = typeof parent == 'string' ? parent : FS.getPath(parent);
		var parts = path.split('/').reverse();
		while (parts.length) {
			var part = parts.pop();
			if (!part) continue;
			var current = PATH.join2(parent, part);
			try {
				FS.mkdir(current);
			} catch (e) {}
			parent = current;
		}
		return current;
	},
	createFile(parent, name, properties, canRead, canWrite) {
		var path = PATH.join2(typeof parent == 'string' ? parent : FS.getPath(parent), name);
		var mode = FS_getMode(canRead, canWrite);
		return FS.create(path, mode);
	},
	createDataFile(parent, name, data, canRead, canWrite, canOwn) {
		var path = name;
		if (parent) {
			parent = typeof parent == 'string' ? parent : FS.getPath(parent);
			path = name ? PATH.join2(parent, name) : parent;
		}
		var mode = FS_getMode(canRead, canWrite);
		var node = FS.create(path, mode);
		if (data) {
			if (typeof data == 'string') {
				var arr = new Array(data.length);
				for (var i = 0, len = data.length; i < len; ++i) arr[i] = data.charCodeAt(i);
				data = arr;
			}
			FS.chmod(node, mode | 146);
			var stream = FS.open(node, 577);
			FS.write(stream, data, 0, data.length, 0, canOwn);
			FS.close(stream);
			FS.chmod(node, mode);
		}
	},
	createDevice(parent, name, input, output) {
		var path = PATH.join2(typeof parent == 'string' ? parent : FS.getPath(parent), name);
		var mode = FS_getMode(!!input, !!output);
		if (!FS.createDevice.major) FS.createDevice.major = 64;
		var dev = FS.makedev(FS.createDevice.major++, 0);
		FS.registerDevice(dev, {
			open(stream) {
				stream.seekable = false;
			},
			close(stream) {
				if (output?.buffer?.length) {
					output(10);
				}
			},
			read(stream, buffer, offset, length, pos) {
				var bytesRead = 0;
				for (var i = 0; i < length; i++) {
					var result;
					try {
						result = input();
					} catch (e) {
						throw new FS.ErrnoError(29);
					}
					if (result === undefined && bytesRead === 0) {
						throw new FS.ErrnoError(6);
					}
					if (result === null || result === undefined) break;
					bytesRead++;
					buffer[offset + i] = result;
				}
				if (bytesRead) {
					stream.node.timestamp = Date.now();
				}
				return bytesRead;
			},
			write(stream, buffer, offset, length, pos) {
				for (var i = 0; i < length; i++) {
					try {
						output(buffer[offset + i]);
					} catch (e) {
						throw new FS.ErrnoError(29);
					}
				}
				if (length) {
					stream.node.timestamp = Date.now();
				}
				return i;
			}
		});
		return FS.mkdev(path, mode, dev);
	},
	forceLoadFile(obj) {
		if (obj.isDevice || obj.isFolder || obj.link || obj.contents) return true;
		if (typeof XMLHttpRequest != 'undefined') {
			throw new Error(
				'Lazy loading should have been performed (contents set) in createLazyFile, but it was not. Lazy loading only works in web workers. Use --embed-file or --preload-file in emcc on the main thread.'
			);
		} else if (read_) {
			try {
				obj.contents = intArrayFromString(read_(obj.url), true);
				obj.usedBytes = obj.contents.length;
			} catch (e) {
				throw new FS.ErrnoError(29);
			}
		} else {
			throw new Error('Cannot load without read() or XMLHttpRequest.');
		}
	},
	createLazyFile(parent, name, url, canRead, canWrite) {
		function LazyUint8Array() {
			this.lengthKnown = false;
			this.chunks = [];
		}
		LazyUint8Array.prototype.get = function LazyUint8Array_get(idx) {
			if (idx > this.length - 1 || idx < 0) {
				return undefined;
			}
			var chunkOffset = idx % this.chunkSize;
			var chunkNum = (idx / this.chunkSize) | 0;
			return this.getter(chunkNum)[chunkOffset];
		};
		LazyUint8Array.prototype.setDataGetter = function LazyUint8Array_setDataGetter(getter) {
			this.getter = getter;
		};
		LazyUint8Array.prototype.cacheLength = function LazyUint8Array_cacheLength() {
			var xhr = new XMLHttpRequest();
			xhr.open('HEAD', url, false);
			xhr.send(null);
			if (!((xhr.status >= 200 && xhr.status < 300) || xhr.status === 304))
				throw new Error("Couldn't load " + url + '. Status: ' + xhr.status);
			var datalength = Number(xhr.getResponseHeader('Content-length'));
			var header;
			var hasByteServing = (header = xhr.getResponseHeader('Accept-Ranges')) && header === 'bytes';
			var usesGzip = (header = xhr.getResponseHeader('Content-Encoding')) && header === 'gzip';
			var chunkSize = 1024 * 1024;
			if (!hasByteServing) chunkSize = datalength;
			var doXHR = (from, to) => {
				if (from > to)
					throw new Error('invalid range (' + from + ', ' + to + ') or no bytes requested!');
				if (to > datalength - 1)
					throw new Error('only ' + datalength + ' bytes available! programmer error!');
				var xhr = new XMLHttpRequest();
				xhr.open('GET', url, false);
				if (datalength !== chunkSize) xhr.setRequestHeader('Range', 'bytes=' + from + '-' + to);
				xhr.responseType = 'arraybuffer';
				if (xhr.overrideMimeType) {
					xhr.overrideMimeType('text/plain; charset=x-user-defined');
				}
				xhr.send(null);
				if (!((xhr.status >= 200 && xhr.status < 300) || xhr.status === 304))
					throw new Error("Couldn't load " + url + '. Status: ' + xhr.status);
				if (xhr.response !== undefined) {
					return new Uint8Array(xhr.response || []);
				}
				return intArrayFromString(xhr.responseText || '', true);
			};
			var lazyArray = this;
			lazyArray.setDataGetter((chunkNum) => {
				var start = chunkNum * chunkSize;
				var end = (chunkNum + 1) * chunkSize - 1;
				end = Math.min(end, datalength - 1);
				if (typeof lazyArray.chunks[chunkNum] == 'undefined') {
					lazyArray.chunks[chunkNum] = doXHR(start, end);
				}
				if (typeof lazyArray.chunks[chunkNum] == 'undefined') throw new Error('doXHR failed!');
				return lazyArray.chunks[chunkNum];
			});
			if (usesGzip || !datalength) {
				chunkSize = datalength = 1;
				datalength = this.getter(0).length;
				chunkSize = datalength;
				out('LazyFiles on gzip forces download of the whole file when length is accessed');
			}
			this._length = datalength;
			this._chunkSize = chunkSize;
			this.lengthKnown = true;
		};
		if (typeof XMLHttpRequest != 'undefined') {
			if (!ENVIRONMENT_IS_WORKER)
				throw 'Cannot do synchronous binary XHRs outside webworkers in modern browsers. Use --embed-file or --preload-file in emcc';
			var lazyArray = new LazyUint8Array();
			Object.defineProperties(lazyArray, {
				length: {
					get: function () {
						if (!this.lengthKnown) {
							this.cacheLength();
						}
						return this._length;
					}
				},
				chunkSize: {
					get: function () {
						if (!this.lengthKnown) {
							this.cacheLength();
						}
						return this._chunkSize;
					}
				}
			});
			var properties = { isDevice: false, contents: lazyArray };
		} else {
			var properties = { isDevice: false, url: url };
		}
		var node = FS.createFile(parent, name, properties, canRead, canWrite);
		if (properties.contents) {
			node.contents = properties.contents;
		} else if (properties.url) {
			node.contents = null;
			node.url = properties.url;
		}
		Object.defineProperties(node, {
			usedBytes: {
				get: function () {
					return this.contents.length;
				}
			}
		});
		var stream_ops = {};
		var keys = Object.keys(node.stream_ops);
		keys.forEach((key) => {
			var fn = node.stream_ops[key];
			stream_ops[key] = function forceLoadLazyFile() {
				FS.forceLoadFile(node);
				return fn.apply(null, arguments);
			};
		});
		function writeChunks(stream, buffer, offset, length, position) {
			var contents = stream.node.contents;
			if (position >= contents.length) return 0;
			var size = Math.min(contents.length - position, length);
			assert(size >= 0);
			if (contents.slice) {
				for (var i = 0; i < size; i++) {
					buffer[offset + i] = contents[position + i];
				}
			} else {
				for (var i = 0; i < size; i++) {
					buffer[offset + i] = contents.get(position + i);
				}
			}
			return size;
		}
		stream_ops.read = (stream, buffer, offset, length, position) => {
			FS.forceLoadFile(node);
			return writeChunks(stream, buffer, offset, length, position);
		};
		stream_ops.mmap = (stream, length, position, prot, flags) => {
			FS.forceLoadFile(node);
			var ptr = mmapAlloc(length);
			if (!ptr) {
				throw new FS.ErrnoError(48);
			}
			writeChunks(stream, HEAP8, ptr, length, position);
			return { ptr: ptr, allocated: true };
		};
		node.stream_ops = stream_ops;
		return node;
	},
	absolutePath() {
		abort('FS.absolutePath has been removed; use PATH_FS.resolve instead');
	},
	createFolder() {
		abort('FS.createFolder has been removed; use FS.mkdir instead');
	},
	createLink() {
		abort('FS.createLink has been removed; use FS.symlink instead');
	},
	joinPath() {
		abort('FS.joinPath has been removed; use PATH.join instead');
	},
	mmapAlloc() {
		abort('FS.mmapAlloc has been replaced by the top level function mmapAlloc');
	},
	standardizePath() {
		abort('FS.standardizePath has been removed; use PATH.normalize instead');
	}
};
var UTF8ToString = (ptr, maxBytesToRead) => {
	assert(typeof ptr == 'number', `UTF8ToString expects a number (got ${typeof ptr})`);
	return ptr ? UTF8ArrayToString(HEAPU8, ptr, maxBytesToRead) : '';
};
var SYSCALLS = {
	DEFAULT_POLLMASK: 5,
	calculateAt(dirfd, path, allowEmpty) {
		if (PATH.isAbs(path)) {
			return path;
		}
		var dir;
		if (dirfd === -100) {
			dir = FS.cwd();
		} else {
			var dirstream = SYSCALLS.getStreamFromFD(dirfd);
			dir = dirstream.path;
		}
		if (path.length == 0) {
			if (!allowEmpty) {
				throw new FS.ErrnoError(44);
			}
			return dir;
		}
		return PATH.join2(dir, path);
	},
	doStat(func, path, buf) {
		try {
			var stat = func(path);
		} catch (e) {
			if (e && e.node && PATH.normalize(path) !== PATH.normalize(FS.getPath(e.node))) {
				return -54;
			}
			throw e;
		}
		HEAP32[buf >> 2] = stat.dev;
		HEAP32[(buf + 4) >> 2] = stat.mode;
		HEAPU32[(buf + 8) >> 2] = stat.nlink;
		HEAP32[(buf + 12) >> 2] = stat.uid;
		HEAP32[(buf + 16) >> 2] = stat.gid;
		HEAP32[(buf + 20) >> 2] = stat.rdev;
		(tempI64 = [
			stat.size >>> 0,
			((tempDouble = stat.size),
			+Math.abs(tempDouble) >= 1
				? tempDouble > 0
					? +Math.floor(tempDouble / 4294967296) >>> 0
					: ~~+Math.ceil((tempDouble - +(~~tempDouble >>> 0)) / 4294967296) >>> 0
				: 0)
		]),
			(HEAP32[(buf + 24) >> 2] = tempI64[0]),
			(HEAP32[(buf + 28) >> 2] = tempI64[1]);
		HEAP32[(buf + 32) >> 2] = 4096;
		HEAP32[(buf + 36) >> 2] = stat.blocks;
		var atime = stat.atime.getTime();
		var mtime = stat.mtime.getTime();
		var ctime = stat.ctime.getTime();
		(tempI64 = [
			Math.floor(atime / 1e3) >>> 0,
			((tempDouble = Math.floor(atime / 1e3)),
			+Math.abs(tempDouble) >= 1
				? tempDouble > 0
					? +Math.floor(tempDouble / 4294967296) >>> 0
					: ~~+Math.ceil((tempDouble - +(~~tempDouble >>> 0)) / 4294967296) >>> 0
				: 0)
		]),
			(HEAP32[(buf + 40) >> 2] = tempI64[0]),
			(HEAP32[(buf + 44) >> 2] = tempI64[1]);
		HEAPU32[(buf + 48) >> 2] = (atime % 1e3) * 1e3;
		(tempI64 = [
			Math.floor(mtime / 1e3) >>> 0,
			((tempDouble = Math.floor(mtime / 1e3)),
			+Math.abs(tempDouble) >= 1
				? tempDouble > 0
					? +Math.floor(tempDouble / 4294967296) >>> 0
					: ~~+Math.ceil((tempDouble - +(~~tempDouble >>> 0)) / 4294967296) >>> 0
				: 0)
		]),
			(HEAP32[(buf + 56) >> 2] = tempI64[0]),
			(HEAP32[(buf + 60) >> 2] = tempI64[1]);
		HEAPU32[(buf + 64) >> 2] = (mtime % 1e3) * 1e3;
		(tempI64 = [
			Math.floor(ctime / 1e3) >>> 0,
			((tempDouble = Math.floor(ctime / 1e3)),
			+Math.abs(tempDouble) >= 1
				? tempDouble > 0
					? +Math.floor(tempDouble / 4294967296) >>> 0
					: ~~+Math.ceil((tempDouble - +(~~tempDouble >>> 0)) / 4294967296) >>> 0
				: 0)
		]),
			(HEAP32[(buf + 72) >> 2] = tempI64[0]),
			(HEAP32[(buf + 76) >> 2] = tempI64[1]);
		HEAPU32[(buf + 80) >> 2] = (ctime % 1e3) * 1e3;
		(tempI64 = [
			stat.ino >>> 0,
			((tempDouble = stat.ino),
			+Math.abs(tempDouble) >= 1
				? tempDouble > 0
					? +Math.floor(tempDouble / 4294967296) >>> 0
					: ~~+Math.ceil((tempDouble - +(~~tempDouble >>> 0)) / 4294967296) >>> 0
				: 0)
		]),
			(HEAP32[(buf + 88) >> 2] = tempI64[0]),
			(HEAP32[(buf + 92) >> 2] = tempI64[1]);
		return 0;
	},
	doMsync(addr, stream, len, flags, offset) {
		if (!FS.isFile(stream.node.mode)) {
			throw new FS.ErrnoError(43);
		}
		if (flags & 2) {
			return 0;
		}
		var buffer = HEAPU8.slice(addr, addr + len);
		FS.msync(stream, buffer, offset, len, flags);
	},
	varargs: undefined,
	get() {
		assert(SYSCALLS.varargs != undefined);
		var ret = HEAP32[+SYSCALLS.varargs >> 2];
		SYSCALLS.varargs += 4;
		return ret;
	},
	getp() {
		return SYSCALLS.get();
	},
	getStr(ptr) {
		var ret = UTF8ToString(ptr);
		return ret;
	},
	getStreamFromFD(fd) {
		var stream = FS.getStreamChecked(fd);
		return stream;
	}
};
var withStackSave = (f) => {
	var stack = stackSave();
	var ret = f();
	stackRestore(stack);
	return ret;
};
var convertI32PairToI53Checked = (lo, hi) => {
	assert(lo == lo >>> 0 || lo == (lo | 0));
	assert(hi === (hi | 0));
	return (hi + 2097152) >>> 0 < 4194305 - !!lo ? (lo >>> 0) + hi * 4294967296 : NaN;
};
var proxyToMainThread = function (index, sync) {
	var numCallArgs = arguments.length - 2;
	var outerArgs = arguments;
	return withStackSave(() => {
		var serializedNumCallArgs = numCallArgs;
		var args = stackAlloc(serializedNumCallArgs * 8);
		var b = args >> 3;
		for (var i = 0; i < numCallArgs; i++) {
			var arg = outerArgs[2 + i];
			HEAPF64[b + i] = arg;
		}
		return __emscripten_run_on_main_thread_js(index, serializedNumCallArgs, args, sync);
	});
};
function _proc_exit(code) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(0, 1, code);
	EXITSTATUS = code;
	if (!keepRuntimeAlive()) {
		PThread.terminateAllThreads();
		Module['onExit']?.(code);
		ABORT = true;
	}
	quit_(code, new ExitStatus(code));
}
var exitJS = (status, implicit) => {
	EXITSTATUS = status;
	checkUnflushedContent();
	if (ENVIRONMENT_IS_PTHREAD) {
		assert(!implicit);
		exitOnMainThread(status);
		throw 'unwind';
	}
	if (keepRuntimeAlive() && !implicit) {
		var msg = `program exited (with status: ${status}), but keepRuntimeAlive() is set (counter=${runtimeKeepaliveCounter}) due to an async operation, so halting execution but not exiting the runtime or preventing further async execution (you can use emscripten_force_exit, if you want to force a true shutdown)`;
		err(msg);
	}
	_proc_exit(status);
};
var _exit = exitJS;
var ptrToString = (ptr) => {
	assert(typeof ptr === 'number');
	ptr >>>= 0;
	return '0x' + ptr.toString(16).padStart(8, '0');
};
var handleException = (e) => {
	if (e instanceof ExitStatus || e == 'unwind') {
		return EXITSTATUS;
	}
	checkStackCookie();
	if (e instanceof WebAssembly.RuntimeError) {
		if (_emscripten_stack_get_current() <= 0) {
			err('Stack overflow detected.  You can try increasing -sSTACK_SIZE (currently set to 65536)');
		}
	}
	quit_(1, e);
};
var PThread = {
	unusedWorkers: [],
	runningWorkers: [],
	tlsInitFunctions: [],
	pthreads: {},
	nextWorkerID: 1,
	debugInit() {
		function pthreadLogPrefix() {
			var t = 0;
			if (runtimeInitialized && typeof _pthread_self != 'undefined') {
				t = _pthread_self();
			}
			return 'w:' + (Module['workerID'] || 0) + ',t:' + ptrToString(t) + ': ';
		}
		var origDbg = dbg;
		dbg = (message) => origDbg(pthreadLogPrefix() + message);
	},
	init() {
		PThread.debugInit();
		if (ENVIRONMENT_IS_PTHREAD) {
			PThread.initWorker();
		} else {
			PThread.initMainThread();
		}
	},
	initMainThread() {
		var pthreadPoolSize = navigator.hardwareConcurrency;
		while (pthreadPoolSize--) {
			PThread.allocateUnusedWorker();
		}
		addOnPreRun(() => {
			addRunDependency('loading-workers');
			PThread.loadWasmModuleToAllWorkers(() => removeRunDependency('loading-workers'));
		});
	},
	initWorker() {
		noExitRuntime = false;
	},
	setExitStatus: (status) => (EXITSTATUS = status),
	terminateAllThreads__deps: ['$terminateWorker'],
	terminateAllThreads: () => {
		assert(
			!ENVIRONMENT_IS_PTHREAD,
			'Internal Error! terminateAllThreads() can only ever be called from main application thread!'
		);
		for (var worker of PThread.runningWorkers) {
			terminateWorker(worker);
		}
		for (var worker of PThread.unusedWorkers) {
			terminateWorker(worker);
		}
		PThread.unusedWorkers = [];
		PThread.runningWorkers = [];
		PThread.pthreads = [];
	},
	returnWorkerToPool: (worker) => {
		var pthread_ptr = worker.pthread_ptr;
		delete PThread.pthreads[pthread_ptr];
		PThread.unusedWorkers.push(worker);
		PThread.runningWorkers.splice(PThread.runningWorkers.indexOf(worker), 1);
		worker.pthread_ptr = 0;
		__emscripten_thread_free_data(pthread_ptr);
	},
	receiveObjectTransfer(data) {},
	threadInitTLS() {
		PThread.tlsInitFunctions.forEach((f) => f());
	},
	loadWasmModuleToWorker: (worker) =>
		new Promise((onFinishedLoading) => {
			worker.onmessage = (e) => {
				var d = e['data'];
				var cmd = d['cmd'];
				if (d['targetThread'] && d['targetThread'] != _pthread_self()) {
					var targetWorker = PThread.pthreads[d['targetThread']];
					if (targetWorker) {
						targetWorker.postMessage(d, d['transferList']);
					} else {
						err(
							`Internal error! Worker sent a message "${cmd}" to target pthread ${d['targetThread']}, but that thread no longer exists!`
						);
					}
					return;
				}
				if (cmd === 'checkMailbox') {
					checkMailbox();
				} else if (cmd === 'spawnThread') {
					spawnThread(d);
				} else if (cmd === 'cleanupThread') {
					cleanupThread(d['thread']);
				} else if (cmd === 'killThread') {
					killThread(d['thread']);
				} else if (cmd === 'cancelThread') {
					cancelThread(d['thread']);
				} else if (cmd === 'loaded') {
					worker.loaded = true;
					if (ENVIRONMENT_IS_NODE && !worker.pthread_ptr) {
						worker.unref();
					}
					onFinishedLoading(worker);
				} else if (cmd === 'alert') {
					alert(`Thread ${d['threadId']}: ${d['text']}`);
				} else if (d.target === 'setimmediate') {
					worker.postMessage(d);
				} else if (cmd === 'callHandler') {
					Module[d['handler']](...d['args']);
				} else if (cmd) {
					err(`worker sent an unknown command ${cmd}`);
				}
			};
			worker.onerror = (e) => {
				var message = 'worker sent an error!';
				if (worker.pthread_ptr) {
					message = `Pthread ${ptrToString(worker.pthread_ptr)} sent an error!`;
				}
				err(`${message} ${e.filename}:${e.lineno}: ${e.message}`);
				throw e;
			};
			if (ENVIRONMENT_IS_NODE) {
				worker.on('message', (data) => worker.onmessage({ data: data }));
				worker.on('error', (e) => worker.onerror(e));
			}
			assert(
				wasmMemory instanceof WebAssembly.Memory,
				'WebAssembly memory should have been loaded by now!'
			);
			assert(
				wasmModule instanceof WebAssembly.Module,
				'WebAssembly Module should have been loaded by now!'
			);
			var handlers = [];
			var knownHandlers = ['onExit', 'onAbort', 'print', 'printErr'];
			for (var handler of knownHandlers) {
				if (Module.hasOwnProperty(handler)) {
					handlers.push(handler);
				}
			}
			worker.workerID = PThread.nextWorkerID++;
			worker.postMessage({
				cmd: 'load',
				handlers: handlers,
				urlOrBlob: Module['mainScriptUrlOrBlob'] || _scriptDir,
				wasmMemory: wasmMemory,
				wasmModule: wasmModule,
				workerID: worker.workerID
			});
		}),
	loadWasmModuleToAllWorkers(onMaybeReady) {
		if (ENVIRONMENT_IS_PTHREAD) {
			return onMaybeReady();
		}
		let pthreadPoolReady = Promise.all(PThread.unusedWorkers.map(PThread.loadWasmModuleToWorker));
		pthreadPoolReady.then(onMaybeReady);
	},
	allocateUnusedWorker() {
		var worker;
		var pthreadMainJs = locateFile('barnes-hut.worker.js');
		worker = new Worker(pthreadMainJs);
		PThread.unusedWorkers.push(worker);
	},
	getNewWorker() {
		if (PThread.unusedWorkers.length == 0) {
			if (!ENVIRONMENT_IS_NODE) {
				err(
					'Tried to spawn a new thread, but the thread pool is exhausted.\n' +
						'This might result in a deadlock unless some threads eventually exit or the code explicitly breaks out to the event loop.\n' +
						'If you want to increase the pool size, use setting `-sPTHREAD_POOL_SIZE=...`.' +
						'\nIf you want to throw an explicit error instead of the risk of deadlocking in those cases, use setting `-sPTHREAD_POOL_SIZE_STRICT=2`.'
				);
			}
			PThread.allocateUnusedWorker();
			PThread.loadWasmModuleToWorker(PThread.unusedWorkers[0]);
		}
		return PThread.unusedWorkers.pop();
	}
};
Module['PThread'] = PThread;
var callRuntimeCallbacks = (callbacks) => {
	while (callbacks.length > 0) {
		callbacks.shift()(Module);
	}
};
var establishStackSpace = () => {
	var pthread_ptr = _pthread_self();
	var stackHigh = HEAPU32[(pthread_ptr + 52) >> 2];
	var stackSize = HEAPU32[(pthread_ptr + 56) >> 2];
	var stackLow = stackHigh - stackSize;
	assert(stackHigh != 0);
	assert(stackLow != 0);
	assert(stackHigh > stackLow, 'stackHigh must be higher then stackLow');
	_emscripten_stack_set_limits(stackHigh, stackLow);
	stackRestore(stackHigh);
	writeStackCookie();
};
Module['establishStackSpace'] = establishStackSpace;
function exitOnMainThread(returnCode) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(1, 0, returnCode);
	_exit(returnCode);
}
var invokeEntryPoint = (ptr, arg) => {
	var result = ((a1) => dynCall_ii.apply(null, [ptr, a1]))(arg);
	checkStackCookie();
	function finish(result) {
		if (keepRuntimeAlive()) {
			PThread.setExitStatus(result);
		} else {
			__emscripten_thread_exit(result);
		}
	}
	finish(result);
};
Module['invokeEntryPoint'] = invokeEntryPoint;
var noExitRuntime = Module['noExitRuntime'] || true;
var registerTLSInit = (tlsInitFunc) => PThread.tlsInitFunctions.push(tlsInitFunc);
var warnOnce = (text) => {
	warnOnce.shown ||= {};
	if (!warnOnce.shown[text]) {
		warnOnce.shown[text] = 1;
		if (ENVIRONMENT_IS_NODE) text = 'warning: ' + text;
		err(text);
	}
};
var ___assert_fail = (condition, filename, line, func) => {
	abort(
		`Assertion failed: ${UTF8ToString(condition)}, at: ` +
			[
				filename ? UTF8ToString(filename) : 'unknown filename',
				line,
				func ? UTF8ToString(func) : 'unknown function'
			]
	);
};
function ExceptionInfo(excPtr) {
	this.excPtr = excPtr;
	this.ptr = excPtr - 24;
	this.set_type = function (type) {
		HEAPU32[(this.ptr + 4) >> 2] = type;
	};
	this.get_type = function () {
		return HEAPU32[(this.ptr + 4) >> 2];
	};
	this.set_destructor = function (destructor) {
		HEAPU32[(this.ptr + 8) >> 2] = destructor;
	};
	this.get_destructor = function () {
		return HEAPU32[(this.ptr + 8) >> 2];
	};
	this.set_caught = function (caught) {
		caught = caught ? 1 : 0;
		HEAP8[(this.ptr + 12) >> 0] = caught;
	};
	this.get_caught = function () {
		return HEAP8[(this.ptr + 12) >> 0] != 0;
	};
	this.set_rethrown = function (rethrown) {
		rethrown = rethrown ? 1 : 0;
		HEAP8[(this.ptr + 13) >> 0] = rethrown;
	};
	this.get_rethrown = function () {
		return HEAP8[(this.ptr + 13) >> 0] != 0;
	};
	this.init = function (type, destructor) {
		this.set_adjusted_ptr(0);
		this.set_type(type);
		this.set_destructor(destructor);
	};
	this.set_adjusted_ptr = function (adjustedPtr) {
		HEAPU32[(this.ptr + 16) >> 2] = adjustedPtr;
	};
	this.get_adjusted_ptr = function () {
		return HEAPU32[(this.ptr + 16) >> 2];
	};
	this.get_exception_ptr = function () {
		var isPointer = ___cxa_is_pointer_type(this.get_type());
		if (isPointer) {
			return HEAPU32[this.excPtr >> 2];
		}
		var adjusted = this.get_adjusted_ptr();
		if (adjusted !== 0) return adjusted;
		return this.excPtr;
	};
}
var exceptionLast = 0;
var uncaughtExceptionCount = 0;
var ___cxa_throw = (ptr, type, destructor) => {
	var info = new ExceptionInfo(ptr);
	info.init(type, destructor);
	exceptionLast = ptr;
	uncaughtExceptionCount++;
	assert(
		false,
		'Exception thrown, but exception catching is not enabled. Compile with -sNO_DISABLE_EXCEPTION_CATCHING or -sEXCEPTION_CATCHING_ALLOWED=[..] to catch.'
	);
};
var ___emscripten_init_main_thread_js = (tb) => {
	__emscripten_thread_init(tb, !ENVIRONMENT_IS_WORKER, 1, !ENVIRONMENT_IS_WEB, 65536, false);
	PThread.threadInitTLS();
};
var ___emscripten_thread_cleanup = (thread) => {
	if (!ENVIRONMENT_IS_PTHREAD) cleanupThread(thread);
	else postMessage({ cmd: 'cleanupThread', thread: thread });
};
function pthreadCreateProxied(pthread_ptr, attr, startRoutine, arg) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(2, 1, pthread_ptr, attr, startRoutine, arg);
	return ___pthread_create_js(pthread_ptr, attr, startRoutine, arg);
}
var ___pthread_create_js = (pthread_ptr, attr, startRoutine, arg) => {
	if (typeof SharedArrayBuffer == 'undefined') {
		err('Current environment does not support SharedArrayBuffer, pthreads are not available!');
		return 6;
	}
	var transferList = [];
	var error = 0;
	if (ENVIRONMENT_IS_PTHREAD && (transferList.length === 0 || error)) {
		return pthreadCreateProxied(pthread_ptr, attr, startRoutine, arg);
	}
	if (error) return error;
	var threadParams = {
		startRoutine: startRoutine,
		pthread_ptr: pthread_ptr,
		arg: arg,
		transferList: transferList
	};
	if (ENVIRONMENT_IS_PTHREAD) {
		threadParams.cmd = 'spawnThread';
		postMessage(threadParams, transferList);
		return 0;
	}
	return spawnThread(threadParams);
};
function ___syscall_faccessat(dirfd, path, amode, flags) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(3, 1, dirfd, path, amode, flags);
	try {
		path = SYSCALLS.getStr(path);
		assert(flags === 0);
		path = SYSCALLS.calculateAt(dirfd, path);
		if (amode & ~7) {
			return -28;
		}
		var lookup = FS.lookupPath(path, { follow: true });
		var node = lookup.node;
		if (!node) {
			return -44;
		}
		var perms = '';
		if (amode & 4) perms += 'r';
		if (amode & 2) perms += 'w';
		if (amode & 1) perms += 'x';
		if (perms && FS.nodePermissions(node, perms)) {
			return -2;
		}
		return 0;
	} catch (e) {
		if (typeof FS == 'undefined' || !(e.name === 'ErrnoError')) throw e;
		return -e.errno;
	}
}
var setErrNo = (value) => {
	HEAP32[___errno_location() >> 2] = value;
	return value;
};
function ___syscall_fcntl64(fd, cmd, varargs) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(4, 1, fd, cmd, varargs);
	SYSCALLS.varargs = varargs;
	try {
		var stream = SYSCALLS.getStreamFromFD(fd);
		switch (cmd) {
			case 0: {
				var arg = SYSCALLS.get();
				if (arg < 0) {
					return -28;
				}
				while (FS.streams[arg]) {
					arg++;
				}
				var newStream;
				newStream = FS.createStream(stream, arg);
				return newStream.fd;
			}
			case 1:
			case 2:
				return 0;
			case 3:
				return stream.flags;
			case 4: {
				var arg = SYSCALLS.get();
				stream.flags |= arg;
				return 0;
			}
			case 5: {
				var arg = SYSCALLS.getp();
				var offset = 0;
				HEAP16[(arg + offset) >> 1] = 2;
				return 0;
			}
			case 6:
			case 7:
				return 0;
			case 16:
			case 8:
				return -28;
			case 9:
				setErrNo(28);
				return -1;
			default: {
				return -28;
			}
		}
	} catch (e) {
		if (typeof FS == 'undefined' || !(e.name === 'ErrnoError')) throw e;
		return -e.errno;
	}
}
var stringToUTF8 = (str, outPtr, maxBytesToWrite) => {
	assert(
		typeof maxBytesToWrite == 'number',
		'stringToUTF8(str, outPtr, maxBytesToWrite) is missing the third parameter that specifies the length of the output buffer!'
	);
	return stringToUTF8Array(str, HEAPU8, outPtr, maxBytesToWrite);
};
function ___syscall_getcwd(buf, size) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(5, 1, buf, size);
	try {
		if (size === 0) return -28;
		var cwd = FS.cwd();
		var cwdLengthInBytes = lengthBytesUTF8(cwd) + 1;
		if (size < cwdLengthInBytes) return -68;
		stringToUTF8(cwd, buf, size);
		return cwdLengthInBytes;
	} catch (e) {
		if (typeof FS == 'undefined' || !(e.name === 'ErrnoError')) throw e;
		return -e.errno;
	}
}
function ___syscall_ioctl(fd, op, varargs) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(6, 1, fd, op, varargs);
	SYSCALLS.varargs = varargs;
	try {
		var stream = SYSCALLS.getStreamFromFD(fd);
		switch (op) {
			case 21509: {
				if (!stream.tty) return -59;
				return 0;
			}
			case 21505: {
				if (!stream.tty) return -59;
				if (stream.tty.ops.ioctl_tcgets) {
					var termios = stream.tty.ops.ioctl_tcgets(stream);
					var argp = SYSCALLS.getp();
					HEAP32[argp >> 2] = termios.c_iflag || 0;
					HEAP32[(argp + 4) >> 2] = termios.c_oflag || 0;
					HEAP32[(argp + 8) >> 2] = termios.c_cflag || 0;
					HEAP32[(argp + 12) >> 2] = termios.c_lflag || 0;
					for (var i = 0; i < 32; i++) {
						HEAP8[(argp + i + 17) >> 0] = termios.c_cc[i] || 0;
					}
					return 0;
				}
				return 0;
			}
			case 21510:
			case 21511:
			case 21512: {
				if (!stream.tty) return -59;
				return 0;
			}
			case 21506:
			case 21507:
			case 21508: {
				if (!stream.tty) return -59;
				if (stream.tty.ops.ioctl_tcsets) {
					var argp = SYSCALLS.getp();
					var c_iflag = HEAP32[argp >> 2];
					var c_oflag = HEAP32[(argp + 4) >> 2];
					var c_cflag = HEAP32[(argp + 8) >> 2];
					var c_lflag = HEAP32[(argp + 12) >> 2];
					var c_cc = [];
					for (var i = 0; i < 32; i++) {
						c_cc.push(HEAP8[(argp + i + 17) >> 0]);
					}
					return stream.tty.ops.ioctl_tcsets(stream.tty, op, {
						c_iflag: c_iflag,
						c_oflag: c_oflag,
						c_cflag: c_cflag,
						c_lflag: c_lflag,
						c_cc: c_cc
					});
				}
				return 0;
			}
			case 21519: {
				if (!stream.tty) return -59;
				var argp = SYSCALLS.getp();
				HEAP32[argp >> 2] = 0;
				return 0;
			}
			case 21520: {
				if (!stream.tty) return -59;
				return -28;
			}
			case 21531: {
				var argp = SYSCALLS.getp();
				return FS.ioctl(stream, op, argp);
			}
			case 21523: {
				if (!stream.tty) return -59;
				if (stream.tty.ops.ioctl_tiocgwinsz) {
					var winsize = stream.tty.ops.ioctl_tiocgwinsz(stream.tty);
					var argp = SYSCALLS.getp();
					HEAP16[argp >> 1] = winsize[0];
					HEAP16[(argp + 2) >> 1] = winsize[1];
				}
				return 0;
			}
			case 21524: {
				if (!stream.tty) return -59;
				return 0;
			}
			case 21515: {
				if (!stream.tty) return -59;
				return 0;
			}
			default:
				return -28;
		}
	} catch (e) {
		if (typeof FS == 'undefined' || !(e.name === 'ErrnoError')) throw e;
		return -e.errno;
	}
}
function ___syscall_openat(dirfd, path, flags, varargs) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(7, 1, dirfd, path, flags, varargs);
	SYSCALLS.varargs = varargs;
	try {
		path = SYSCALLS.getStr(path);
		path = SYSCALLS.calculateAt(dirfd, path);
		var mode = varargs ? SYSCALLS.get() : 0;
		return FS.open(path, flags, mode).fd;
	} catch (e) {
		if (typeof FS == 'undefined' || !(e.name === 'ErrnoError')) throw e;
		return -e.errno;
	}
}
var nowIsMonotonic = 1;
var __emscripten_get_now_is_monotonic = () => nowIsMonotonic;
var maybeExit = () => {
	if (!keepRuntimeAlive()) {
		try {
			if (ENVIRONMENT_IS_PTHREAD) __emscripten_thread_exit(EXITSTATUS);
			else _exit(EXITSTATUS);
		} catch (e) {
			handleException(e);
		}
	}
};
var callUserCallback = (func) => {
	if (ABORT) {
		err('user callback triggered after runtime exited or application aborted.  Ignoring.');
		return;
	}
	try {
		func();
		maybeExit();
	} catch (e) {
		handleException(e);
	}
};
var __emscripten_thread_mailbox_await = (pthread_ptr) => {
	if (typeof Atomics.waitAsync === 'function') {
		var wait = Atomics.waitAsync(HEAP32, pthread_ptr >> 2, pthread_ptr);
		assert(wait.async);
		wait.value.then(checkMailbox);
		var waitingAsync = pthread_ptr + 128;
		Atomics.store(HEAP32, waitingAsync >> 2, 1);
	}
};
Module['__emscripten_thread_mailbox_await'] = __emscripten_thread_mailbox_await;
var checkMailbox = () => {
	var pthread_ptr = _pthread_self();
	if (pthread_ptr) {
		__emscripten_thread_mailbox_await(pthread_ptr);
		callUserCallback(__emscripten_check_mailbox);
	}
};
Module['checkMailbox'] = checkMailbox;
var __emscripten_notify_mailbox_postmessage = (targetThreadId, currThreadId, mainThreadId) => {
	if (targetThreadId == currThreadId) {
		setTimeout(() => checkMailbox());
	} else if (ENVIRONMENT_IS_PTHREAD) {
		postMessage({ targetThread: targetThreadId, cmd: 'checkMailbox' });
	} else {
		var worker = PThread.pthreads[targetThreadId];
		if (!worker) {
			err(`Cannot send message to thread with ID ${targetThreadId}, unknown thread ID!`);
			return;
		}
		worker.postMessage({ cmd: 'checkMailbox' });
	}
};
var proxiedJSCallArgs = [];
var __emscripten_receive_on_main_thread_js = (index, callingThread, numCallArgs, args) => {
	proxiedJSCallArgs.length = numCallArgs;
	var b = args >> 3;
	for (var i = 0; i < numCallArgs; i++) {
		proxiedJSCallArgs[i] = HEAPF64[b + i];
	}
	var isEmAsmConst = index < 0;
	var func = !isEmAsmConst ? proxiedFunctionTable[index] : ASM_CONSTS[-index - 1];
	assert(func.length == numCallArgs, 'Call args mismatch in _emscripten_receive_on_main_thread_js');
	PThread.currentProxiedOperationCallerThread = callingThread;
	var rtn = func.apply(null, proxiedJSCallArgs);
	PThread.currentProxiedOperationCallerThread = 0;
	assert(typeof rtn != 'bigint');
	return rtn;
};
var __emscripten_thread_set_strongref = (thread) => {
	if (ENVIRONMENT_IS_NODE) {
		PThread.pthreads[thread].ref();
	}
};
var _abort = () => {
	abort('native code called abort()');
};
var readEmAsmArgsArray = [];
var readEmAsmArgs = (sigPtr, buf) => {
	assert(Array.isArray(readEmAsmArgsArray));
	assert(buf % 16 == 0);
	readEmAsmArgsArray.length = 0;
	var ch;
	while ((ch = HEAPU8[sigPtr++])) {
		var chr = String.fromCharCode(ch);
		var validChars = ['d', 'f', 'i', 'p'];
		assert(
			validChars.includes(chr),
			`Invalid character ${ch}("${chr}") in readEmAsmArgs! Use only [${validChars}], and do not specify "v" for void return argument.`
		);
		var wide = ch != 105;
		wide &= ch != 112;
		buf += wide && buf % 8 ? 4 : 0;
		readEmAsmArgsArray.push(
			ch == 112 ? HEAPU32[buf >> 2] : ch == 105 ? HEAP32[buf >> 2] : HEAPF64[buf >> 3]
		);
		buf += wide ? 8 : 4;
	}
	return readEmAsmArgsArray;
};
var runEmAsmFunction = (code, sigPtr, argbuf) => {
	var args = readEmAsmArgs(sigPtr, argbuf);
	assert(
		ASM_CONSTS.hasOwnProperty(code),
		`No EM_ASM constant found at address ${code}.  The loaded WebAssembly file is likely out of sync with the generated JavaScript.`
	);
	return ASM_CONSTS[code].apply(null, args);
};
var _emscripten_asm_const_int = (code, sigPtr, argbuf) => runEmAsmFunction(code, sigPtr, argbuf);
var _emscripten_check_blocking_allowed = () => {
	if (ENVIRONMENT_IS_NODE) return;
	if (ENVIRONMENT_IS_WORKER) return;
	warnOnce(
		'Blocking on the main thread is very dangerous, see https://emscripten.org/docs/porting/pthreads.html#blocking-on-the-main-browser-thread'
	);
};
var _emscripten_date_now = () => Date.now();
var runtimeKeepalivePush = () => {
	runtimeKeepaliveCounter += 1;
};
var _emscripten_exit_with_live_runtime = () => {
	runtimeKeepalivePush();
	throw 'unwind';
};
var JSEvents = {
	inEventHandler: 0,
	removeAllEventListeners() {
		for (var i = JSEvents.eventHandlers.length - 1; i >= 0; --i) {
			JSEvents._removeHandler(i);
		}
		JSEvents.eventHandlers = [];
		JSEvents.deferredCalls = [];
	},
	registerRemoveEventListeners() {
		if (!JSEvents.removeEventListenersRegistered) {
			__ATEXIT__.push(JSEvents.removeAllEventListeners);
			JSEvents.removeEventListenersRegistered = true;
		}
	},
	deferredCalls: [],
	deferCall(targetFunction, precedence, argsList) {
		function arraysHaveEqualContent(arrA, arrB) {
			if (arrA.length != arrB.length) return false;
			for (var i in arrA) {
				if (arrA[i] != arrB[i]) return false;
			}
			return true;
		}
		for (var i in JSEvents.deferredCalls) {
			var call = JSEvents.deferredCalls[i];
			if (
				call.targetFunction == targetFunction &&
				arraysHaveEqualContent(call.argsList, argsList)
			) {
				return;
			}
		}
		JSEvents.deferredCalls.push({
			targetFunction: targetFunction,
			precedence: precedence,
			argsList: argsList
		});
		JSEvents.deferredCalls.sort((x, y) => x.precedence < y.precedence);
	},
	removeDeferredCalls(targetFunction) {
		for (var i = 0; i < JSEvents.deferredCalls.length; ++i) {
			if (JSEvents.deferredCalls[i].targetFunction == targetFunction) {
				JSEvents.deferredCalls.splice(i, 1);
				--i;
			}
		}
	},
	canPerformEventHandlerRequests() {
		if (navigator.userActivation) {
			return navigator.userActivation.isActive;
		}
		return JSEvents.inEventHandler && JSEvents.currentEventHandler.allowsDeferredCalls;
	},
	runDeferredCalls() {
		if (!JSEvents.canPerformEventHandlerRequests()) {
			return;
		}
		for (var i = 0; i < JSEvents.deferredCalls.length; ++i) {
			var call = JSEvents.deferredCalls[i];
			JSEvents.deferredCalls.splice(i, 1);
			--i;
			call.targetFunction.apply(null, call.argsList);
		}
	},
	eventHandlers: [],
	removeAllHandlersOnTarget: (target, eventTypeString) => {
		for (var i = 0; i < JSEvents.eventHandlers.length; ++i) {
			if (
				JSEvents.eventHandlers[i].target == target &&
				(!eventTypeString || eventTypeString == JSEvents.eventHandlers[i].eventTypeString)
			) {
				JSEvents._removeHandler(i--);
			}
		}
	},
	_removeHandler(i) {
		var h = JSEvents.eventHandlers[i];
		h.target.removeEventListener(h.eventTypeString, h.eventListenerFunc, h.useCapture);
		JSEvents.eventHandlers.splice(i, 1);
	},
	registerOrRemoveHandler(eventHandler) {
		if (!eventHandler.target) {
			err(
				'registerOrRemoveHandler: the target element for event handler registration does not exist, when processing the following event handler registration:'
			);
			console.dir(eventHandler);
			return -4;
		}
		var jsEventHandler = function jsEventHandler(event) {
			++JSEvents.inEventHandler;
			JSEvents.currentEventHandler = eventHandler;
			JSEvents.runDeferredCalls();
			eventHandler.handlerFunc(event);
			JSEvents.runDeferredCalls();
			--JSEvents.inEventHandler;
		};
		if (eventHandler.callbackfunc) {
			eventHandler.eventListenerFunc = jsEventHandler;
			eventHandler.target.addEventListener(
				eventHandler.eventTypeString,
				jsEventHandler,
				eventHandler.useCapture
			);
			JSEvents.eventHandlers.push(eventHandler);
			JSEvents.registerRemoveEventListeners();
		} else {
			for (var i = 0; i < JSEvents.eventHandlers.length; ++i) {
				if (
					JSEvents.eventHandlers[i].target == eventHandler.target &&
					JSEvents.eventHandlers[i].eventTypeString == eventHandler.eventTypeString
				) {
					JSEvents._removeHandler(i--);
				}
			}
		}
		return 0;
	},
	getTargetThreadForEventCallback(targetThread) {
		switch (targetThread) {
			case 1:
				return 0;
			case 2:
				return PThread.currentProxiedOperationCallerThread;
			default:
				return targetThread;
		}
	},
	getNodeNameForTarget(target) {
		if (!target) return '';
		if (target == window) return '#window';
		if (target == screen) return '#screen';
		return target?.nodeName || '';
	},
	fullscreenEnabled() {
		return document.fullscreenEnabled || document.webkitFullscreenEnabled;
	}
};
var maybeCStringToJsString = (cString) => (cString > 2 ? UTF8ToString(cString) : cString);
var specialHTMLTargets = [
	0,
	typeof document != 'undefined' ? document : 0,
	typeof window != 'undefined' ? window : 0
];
var findEventTarget = (target) => {
	target = maybeCStringToJsString(target);
	var domElement =
		specialHTMLTargets[target] ||
		(typeof document != 'undefined' ? document.querySelector(target) : undefined);
	return domElement;
};
var getBoundingClientRect = (e) =>
	specialHTMLTargets.indexOf(e) < 0 ? e.getBoundingClientRect() : { left: 0, top: 0 };
function _emscripten_get_element_css_size(target, width, height) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(8, 1, target, width, height);
	target = findEventTarget(target);
	if (!target) return -4;
	var rect = getBoundingClientRect(target);
	HEAPF64[width >> 3] = rect.width;
	HEAPF64[height >> 3] = rect.height;
	return 0;
}
var fillGamepadEventData = (eventStruct, e) => {
	HEAPF64[eventStruct >> 3] = e.timestamp;
	for (var i = 0; i < e.axes.length; ++i) {
		HEAPF64[(eventStruct + i * 8 + 16) >> 3] = e.axes[i];
	}
	for (var i = 0; i < e.buttons.length; ++i) {
		if (typeof e.buttons[i] == 'object') {
			HEAPF64[(eventStruct + i * 8 + 528) >> 3] = e.buttons[i].value;
		} else {
			HEAPF64[(eventStruct + i * 8 + 528) >> 3] = e.buttons[i];
		}
	}
	for (var i = 0; i < e.buttons.length; ++i) {
		if (typeof e.buttons[i] == 'object') {
			HEAP32[(eventStruct + i * 4 + 1040) >> 2] = e.buttons[i].pressed;
		} else {
			HEAP32[(eventStruct + i * 4 + 1040) >> 2] = e.buttons[i] == 1;
		}
	}
	HEAP32[(eventStruct + 1296) >> 2] = e.connected;
	HEAP32[(eventStruct + 1300) >> 2] = e.index;
	HEAP32[(eventStruct + 8) >> 2] = e.axes.length;
	HEAP32[(eventStruct + 12) >> 2] = e.buttons.length;
	stringToUTF8(e.id, eventStruct + 1304, 64);
	stringToUTF8(e.mapping, eventStruct + 1368, 64);
};
function _emscripten_get_gamepad_status(index, gamepadState) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(9, 1, index, gamepadState);
	if (!JSEvents.lastGamepadState)
		throw 'emscripten_get_gamepad_status() can only be called after having first called emscripten_sample_gamepad_data() and that function has returned EMSCRIPTEN_RESULT_SUCCESS!';
	if (index < 0 || index >= JSEvents.lastGamepadState.length) return -5;
	if (!JSEvents.lastGamepadState[index]) return -7;
	fillGamepadEventData(gamepadState, JSEvents.lastGamepadState[index]);
	return 0;
}
var getHeapMax = () => HEAPU8.length;
var _emscripten_get_heap_max = () => getHeapMax();
var _emscripten_get_now;
_emscripten_get_now = () => performance.timeOrigin + performance.now();
function _emscripten_get_num_gamepads() {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(10, 1);
	if (!JSEvents.lastGamepadState)
		throw 'emscripten_get_num_gamepads() can only be called after having first called emscripten_sample_gamepad_data() and that function has returned EMSCRIPTEN_RESULT_SUCCESS!';
	return JSEvents.lastGamepadState.length;
}
var webgl_enable_ANGLE_instanced_arrays = (ctx) => {
	var ext = ctx.getExtension('ANGLE_instanced_arrays');
	if (ext) {
		ctx['vertexAttribDivisor'] = (index, divisor) =>
			ext['vertexAttribDivisorANGLE'](index, divisor);
		ctx['drawArraysInstanced'] = (mode, first, count, primcount) =>
			ext['drawArraysInstancedANGLE'](mode, first, count, primcount);
		ctx['drawElementsInstanced'] = (mode, count, type, indices, primcount) =>
			ext['drawElementsInstancedANGLE'](mode, count, type, indices, primcount);
		return 1;
	}
};
var webgl_enable_OES_vertex_array_object = (ctx) => {
	var ext = ctx.getExtension('OES_vertex_array_object');
	if (ext) {
		ctx['createVertexArray'] = () => ext['createVertexArrayOES']();
		ctx['deleteVertexArray'] = (vao) => ext['deleteVertexArrayOES'](vao);
		ctx['bindVertexArray'] = (vao) => ext['bindVertexArrayOES'](vao);
		ctx['isVertexArray'] = (vao) => ext['isVertexArrayOES'](vao);
		return 1;
	}
};
var webgl_enable_WEBGL_draw_buffers = (ctx) => {
	var ext = ctx.getExtension('WEBGL_draw_buffers');
	if (ext) {
		ctx['drawBuffers'] = (n, bufs) => ext['drawBuffersWEBGL'](n, bufs);
		return 1;
	}
};
var webgl_enable_WEBGL_multi_draw = (ctx) =>
	!!(ctx.multiDrawWebgl = ctx.getExtension('WEBGL_multi_draw'));
var GL = {
	counter: 1,
	buffers: [],
	programs: [],
	framebuffers: [],
	renderbuffers: [],
	textures: [],
	shaders: [],
	vaos: [],
	contexts: {},
	offscreenCanvases: {},
	queries: [],
	stringCache: {},
	unpackAlignment: 4,
	recordError: function recordError(errorCode) {
		if (!GL.lastError) {
			GL.lastError = errorCode;
		}
	},
	getNewId: (table) => {
		var ret = GL.counter++;
		for (var i = table.length; i < ret; i++) {
			table[i] = null;
		}
		return ret;
	},
	getSource: (shader, count, string, length) => {
		var source = '';
		for (var i = 0; i < count; ++i) {
			var len = length ? HEAP32[(length + i * 4) >> 2] : -1;
			source += UTF8ToString(HEAP32[(string + i * 4) >> 2], len < 0 ? undefined : len);
		}
		return source;
	},
	createContext: (canvas, webGLContextAttributes) => {
		if (!canvas.getContextSafariWebGL2Fixed) {
			canvas.getContextSafariWebGL2Fixed = canvas.getContext;
			function fixedGetContext(ver, attrs) {
				var gl = canvas.getContextSafariWebGL2Fixed(ver, attrs);
				return (ver == 'webgl') == gl instanceof WebGLRenderingContext ? gl : null;
			}
			canvas.getContext = fixedGetContext;
		}
		var ctx = canvas.getContext('webgl', webGLContextAttributes);
		if (!ctx) return 0;
		var handle = GL.registerContext(ctx, webGLContextAttributes);
		var _allSupportedExtensions = ctx.getSupportedExtensions;
		var supportedExtensionsForGetProcAddress = [
			'ANGLE_instanced_arrays',
			'EXT_blend_minmax',
			'EXT_disjoint_timer_query',
			'EXT_frag_depth',
			'EXT_shader_texture_lod',
			'EXT_sRGB',
			'OES_element_index_uint',
			'OES_fbo_render_mipmap',
			'OES_standard_derivatives',
			'OES_texture_float',
			'OES_texture_half_float',
			'OES_texture_half_float_linear',
			'OES_vertex_array_object',
			'WEBGL_color_buffer_float',
			'WEBGL_depth_texture',
			'WEBGL_draw_buffers',
			'EXT_color_buffer_half_float',
			'EXT_float_blend',
			'EXT_texture_compression_bptc',
			'EXT_texture_compression_rgtc',
			'EXT_texture_filter_anisotropic',
			'KHR_parallel_shader_compile',
			'OES_texture_float_linear',
			'WEBGL_compressed_texture_s3tc',
			'WEBGL_compressed_texture_s3tc_srgb',
			'WEBGL_debug_renderer_info',
			'WEBGL_debug_shaders',
			'WEBGL_lose_context',
			'WEBGL_multi_draw'
		];
		ctx.getSupportedExtensions = function () {
			return (_allSupportedExtensions.apply(this) || []).filter((ext) =>
				supportedExtensionsForGetProcAddress.includes(ext)
			);
		};
		return handle;
	},
	registerContext: (ctx, webGLContextAttributes) => {
		var handle = _malloc(8);
		HEAPU32[(handle + 4) >> 2] = _pthread_self();
		var context = {
			handle: handle,
			attributes: webGLContextAttributes,
			version: webGLContextAttributes.majorVersion,
			GLctx: ctx
		};
		if (ctx.canvas) ctx.canvas.GLctxObject = context;
		GL.contexts[handle] = context;
		if (
			typeof webGLContextAttributes.enableExtensionsByDefault == 'undefined' ||
			webGLContextAttributes.enableExtensionsByDefault
		) {
			GL.initExtensions(context);
		}
		return handle;
	},
	makeContextCurrent: (contextHandle) => {
		GL.currentContext = GL.contexts[contextHandle];
		Module.ctx = GLctx = GL.currentContext?.GLctx;
		return !(contextHandle && !GLctx);
	},
	getContext: (contextHandle) => GL.contexts[contextHandle],
	deleteContext: (contextHandle) => {
		if (GL.currentContext === GL.contexts[contextHandle]) {
			GL.currentContext = null;
		}
		if (typeof JSEvents == 'object') {
			JSEvents.removeAllHandlersOnTarget(GL.contexts[contextHandle].GLctx.canvas);
		}
		if (GL.contexts[contextHandle] && GL.contexts[contextHandle].GLctx.canvas) {
			GL.contexts[contextHandle].GLctx.canvas.GLctxObject = undefined;
		}
		_free(GL.contexts[contextHandle].handle);
		GL.contexts[contextHandle] = null;
	},
	initExtensions: (context) => {
		context ||= GL.currentContext;
		if (context.initExtensionsDone) return;
		context.initExtensionsDone = true;
		var GLctx = context.GLctx;
		webgl_enable_ANGLE_instanced_arrays(GLctx);
		webgl_enable_OES_vertex_array_object(GLctx);
		webgl_enable_WEBGL_draw_buffers(GLctx);
		{
			GLctx.disjointTimerQueryExt = GLctx.getExtension('EXT_disjoint_timer_query');
		}
		webgl_enable_WEBGL_multi_draw(GLctx);
		var exts = GLctx.getSupportedExtensions() || [];
		exts.forEach((ext) => {
			if (!ext.includes('lose_context') && !ext.includes('debug')) {
				GLctx.getExtension(ext);
			}
		});
	},
	getExtensions() {
		var exts = GLctx.getSupportedExtensions() || [];
		exts = exts.concat(exts.map((e) => 'GL_' + e));
		return exts;
	}
};
function _glActiveTexture(x0) {
	GLctx.activeTexture(x0);
}
var _emscripten_glActiveTexture = _glActiveTexture;
var _glAttachShader = (program, shader) => {
	GLctx.attachShader(GL.programs[program], GL.shaders[shader]);
};
var _emscripten_glAttachShader = _glAttachShader;
var _glBeginQueryEXT = (target, id) => {
	GLctx.disjointTimerQueryExt['beginQueryEXT'](target, GL.queries[id]);
};
var _emscripten_glBeginQueryEXT = _glBeginQueryEXT;
var _glBindAttribLocation = (program, index, name) => {
	GLctx.bindAttribLocation(GL.programs[program], index, UTF8ToString(name));
};
var _emscripten_glBindAttribLocation = _glBindAttribLocation;
var _glBindBuffer = (target, buffer) => {
	GLctx.bindBuffer(target, GL.buffers[buffer]);
};
var _emscripten_glBindBuffer = _glBindBuffer;
var _glBindFramebuffer = (target, framebuffer) => {
	GLctx.bindFramebuffer(target, GL.framebuffers[framebuffer]);
};
var _emscripten_glBindFramebuffer = _glBindFramebuffer;
var _glBindRenderbuffer = (target, renderbuffer) => {
	GLctx.bindRenderbuffer(target, GL.renderbuffers[renderbuffer]);
};
var _emscripten_glBindRenderbuffer = _glBindRenderbuffer;
var _glBindTexture = (target, texture) => {
	GLctx.bindTexture(target, GL.textures[texture]);
};
var _emscripten_glBindTexture = _glBindTexture;
var _glBindVertexArray = (vao) => {
	GLctx.bindVertexArray(GL.vaos[vao]);
};
var _glBindVertexArrayOES = _glBindVertexArray;
var _emscripten_glBindVertexArrayOES = _glBindVertexArrayOES;
function _glBlendColor(x0, x1, x2, x3) {
	GLctx.blendColor(x0, x1, x2, x3);
}
var _emscripten_glBlendColor = _glBlendColor;
function _glBlendEquation(x0) {
	GLctx.blendEquation(x0);
}
var _emscripten_glBlendEquation = _glBlendEquation;
function _glBlendEquationSeparate(x0, x1) {
	GLctx.blendEquationSeparate(x0, x1);
}
var _emscripten_glBlendEquationSeparate = _glBlendEquationSeparate;
function _glBlendFunc(x0, x1) {
	GLctx.blendFunc(x0, x1);
}
var _emscripten_glBlendFunc = _glBlendFunc;
function _glBlendFuncSeparate(x0, x1, x2, x3) {
	GLctx.blendFuncSeparate(x0, x1, x2, x3);
}
var _emscripten_glBlendFuncSeparate = _glBlendFuncSeparate;
var _glBufferData = (target, size, data, usage) => {
	GLctx.bufferData(target, data ? HEAPU8.subarray(data, data + size) : size, usage);
};
var _emscripten_glBufferData = _glBufferData;
var _glBufferSubData = (target, offset, size, data) => {
	GLctx.bufferSubData(target, offset, HEAPU8.subarray(data, data + size));
};
var _emscripten_glBufferSubData = _glBufferSubData;
function _glCheckFramebufferStatus(x0) {
	return GLctx.checkFramebufferStatus(x0);
}
var _emscripten_glCheckFramebufferStatus = _glCheckFramebufferStatus;
function _glClear(x0) {
	GLctx.clear(x0);
}
var _emscripten_glClear = _glClear;
function _glClearColor(x0, x1, x2, x3) {
	GLctx.clearColor(x0, x1, x2, x3);
}
var _emscripten_glClearColor = _glClearColor;
function _glClearDepthf(x0) {
	GLctx.clearDepth(x0);
}
var _emscripten_glClearDepthf = _glClearDepthf;
function _glClearStencil(x0) {
	GLctx.clearStencil(x0);
}
var _emscripten_glClearStencil = _glClearStencil;
var _glColorMask = (red, green, blue, alpha) => {
	GLctx.colorMask(!!red, !!green, !!blue, !!alpha);
};
var _emscripten_glColorMask = _glColorMask;
var _glCompileShader = (shader) => {
	GLctx.compileShader(GL.shaders[shader]);
};
var _emscripten_glCompileShader = _glCompileShader;
var _glCompressedTexImage2D = (
	target,
	level,
	internalFormat,
	width,
	height,
	border,
	imageSize,
	data
) => {
	GLctx.compressedTexImage2D(
		target,
		level,
		internalFormat,
		width,
		height,
		border,
		data ? HEAPU8.subarray(data, data + imageSize) : null
	);
};
var _emscripten_glCompressedTexImage2D = _glCompressedTexImage2D;
var _glCompressedTexSubImage2D = (
	target,
	level,
	xoffset,
	yoffset,
	width,
	height,
	format,
	imageSize,
	data
) => {
	GLctx.compressedTexSubImage2D(
		target,
		level,
		xoffset,
		yoffset,
		width,
		height,
		format,
		data ? HEAPU8.subarray(data, data + imageSize) : null
	);
};
var _emscripten_glCompressedTexSubImage2D = _glCompressedTexSubImage2D;
function _glCopyTexImage2D(x0, x1, x2, x3, x4, x5, x6, x7) {
	GLctx.copyTexImage2D(x0, x1, x2, x3, x4, x5, x6, x7);
}
var _emscripten_glCopyTexImage2D = _glCopyTexImage2D;
function _glCopyTexSubImage2D(x0, x1, x2, x3, x4, x5, x6, x7) {
	GLctx.copyTexSubImage2D(x0, x1, x2, x3, x4, x5, x6, x7);
}
var _emscripten_glCopyTexSubImage2D = _glCopyTexSubImage2D;
var _glCreateProgram = () => {
	var id = GL.getNewId(GL.programs);
	var program = GLctx.createProgram();
	program.name = id;
	program.maxUniformLength = program.maxAttributeLength = program.maxUniformBlockNameLength = 0;
	program.uniformIdCounter = 1;
	GL.programs[id] = program;
	return id;
};
var _emscripten_glCreateProgram = _glCreateProgram;
var _glCreateShader = (shaderType) => {
	var id = GL.getNewId(GL.shaders);
	GL.shaders[id] = GLctx.createShader(shaderType);
	return id;
};
var _emscripten_glCreateShader = _glCreateShader;
function _glCullFace(x0) {
	GLctx.cullFace(x0);
}
var _emscripten_glCullFace = _glCullFace;
var _glDeleteBuffers = (n, buffers) => {
	for (var i = 0; i < n; i++) {
		var id = HEAP32[(buffers + i * 4) >> 2];
		var buffer = GL.buffers[id];
		if (!buffer) continue;
		GLctx.deleteBuffer(buffer);
		buffer.name = 0;
		GL.buffers[id] = null;
	}
};
var _emscripten_glDeleteBuffers = _glDeleteBuffers;
var _glDeleteFramebuffers = (n, framebuffers) => {
	for (var i = 0; i < n; ++i) {
		var id = HEAP32[(framebuffers + i * 4) >> 2];
		var framebuffer = GL.framebuffers[id];
		if (!framebuffer) continue;
		GLctx.deleteFramebuffer(framebuffer);
		framebuffer.name = 0;
		GL.framebuffers[id] = null;
	}
};
var _emscripten_glDeleteFramebuffers = _glDeleteFramebuffers;
var _glDeleteProgram = (id) => {
	if (!id) return;
	var program = GL.programs[id];
	if (!program) {
		GL.recordError(1281);
		return;
	}
	GLctx.deleteProgram(program);
	program.name = 0;
	GL.programs[id] = null;
};
var _emscripten_glDeleteProgram = _glDeleteProgram;
var _glDeleteQueriesEXT = (n, ids) => {
	for (var i = 0; i < n; i++) {
		var id = HEAP32[(ids + i * 4) >> 2];
		var query = GL.queries[id];
		if (!query) continue;
		GLctx.disjointTimerQueryExt['deleteQueryEXT'](query);
		GL.queries[id] = null;
	}
};
var _emscripten_glDeleteQueriesEXT = _glDeleteQueriesEXT;
var _glDeleteRenderbuffers = (n, renderbuffers) => {
	for (var i = 0; i < n; i++) {
		var id = HEAP32[(renderbuffers + i * 4) >> 2];
		var renderbuffer = GL.renderbuffers[id];
		if (!renderbuffer) continue;
		GLctx.deleteRenderbuffer(renderbuffer);
		renderbuffer.name = 0;
		GL.renderbuffers[id] = null;
	}
};
var _emscripten_glDeleteRenderbuffers = _glDeleteRenderbuffers;
var _glDeleteShader = (id) => {
	if (!id) return;
	var shader = GL.shaders[id];
	if (!shader) {
		GL.recordError(1281);
		return;
	}
	GLctx.deleteShader(shader);
	GL.shaders[id] = null;
};
var _emscripten_glDeleteShader = _glDeleteShader;
var _glDeleteTextures = (n, textures) => {
	for (var i = 0; i < n; i++) {
		var id = HEAP32[(textures + i * 4) >> 2];
		var texture = GL.textures[id];
		if (!texture) continue;
		GLctx.deleteTexture(texture);
		texture.name = 0;
		GL.textures[id] = null;
	}
};
var _emscripten_glDeleteTextures = _glDeleteTextures;
var _glDeleteVertexArrays = (n, vaos) => {
	for (var i = 0; i < n; i++) {
		var id = HEAP32[(vaos + i * 4) >> 2];
		GLctx.deleteVertexArray(GL.vaos[id]);
		GL.vaos[id] = null;
	}
};
var _glDeleteVertexArraysOES = _glDeleteVertexArrays;
var _emscripten_glDeleteVertexArraysOES = _glDeleteVertexArraysOES;
function _glDepthFunc(x0) {
	GLctx.depthFunc(x0);
}
var _emscripten_glDepthFunc = _glDepthFunc;
var _glDepthMask = (flag) => {
	GLctx.depthMask(!!flag);
};
var _emscripten_glDepthMask = _glDepthMask;
function _glDepthRangef(x0, x1) {
	GLctx.depthRange(x0, x1);
}
var _emscripten_glDepthRangef = _glDepthRangef;
var _glDetachShader = (program, shader) => {
	GLctx.detachShader(GL.programs[program], GL.shaders[shader]);
};
var _emscripten_glDetachShader = _glDetachShader;
function _glDisable(x0) {
	GLctx.disable(x0);
}
var _emscripten_glDisable = _glDisable;
var _glDisableVertexAttribArray = (index) => {
	GLctx.disableVertexAttribArray(index);
};
var _emscripten_glDisableVertexAttribArray = _glDisableVertexAttribArray;
var _glDrawArrays = (mode, first, count) => {
	GLctx.drawArrays(mode, first, count);
};
var _emscripten_glDrawArrays = _glDrawArrays;
var _glDrawArraysInstanced = (mode, first, count, primcount) => {
	GLctx.drawArraysInstanced(mode, first, count, primcount);
};
var _glDrawArraysInstancedANGLE = _glDrawArraysInstanced;
var _emscripten_glDrawArraysInstancedANGLE = _glDrawArraysInstancedANGLE;
var tempFixedLengthArray = [];
var _glDrawBuffers = (n, bufs) => {
	var bufArray = tempFixedLengthArray[n];
	for (var i = 0; i < n; i++) {
		bufArray[i] = HEAP32[(bufs + i * 4) >> 2];
	}
	GLctx.drawBuffers(bufArray);
};
var _glDrawBuffersWEBGL = _glDrawBuffers;
var _emscripten_glDrawBuffersWEBGL = _glDrawBuffersWEBGL;
var _glDrawElements = (mode, count, type, indices) => {
	GLctx.drawElements(mode, count, type, indices);
};
var _emscripten_glDrawElements = _glDrawElements;
var _glDrawElementsInstanced = (mode, count, type, indices, primcount) => {
	GLctx.drawElementsInstanced(mode, count, type, indices, primcount);
};
var _glDrawElementsInstancedANGLE = _glDrawElementsInstanced;
var _emscripten_glDrawElementsInstancedANGLE = _glDrawElementsInstancedANGLE;
function _glEnable(x0) {
	GLctx.enable(x0);
}
var _emscripten_glEnable = _glEnable;
var _glEnableVertexAttribArray = (index) => {
	GLctx.enableVertexAttribArray(index);
};
var _emscripten_glEnableVertexAttribArray = _glEnableVertexAttribArray;
var _glEndQueryEXT = (target) => {
	GLctx.disjointTimerQueryExt['endQueryEXT'](target);
};
var _emscripten_glEndQueryEXT = _glEndQueryEXT;
function _glFinish() {
	GLctx.finish();
}
var _emscripten_glFinish = _glFinish;
function _glFlush() {
	GLctx.flush();
}
var _emscripten_glFlush = _glFlush;
var _glFramebufferRenderbuffer = (target, attachment, renderbuffertarget, renderbuffer) => {
	GLctx.framebufferRenderbuffer(
		target,
		attachment,
		renderbuffertarget,
		GL.renderbuffers[renderbuffer]
	);
};
var _emscripten_glFramebufferRenderbuffer = _glFramebufferRenderbuffer;
var _glFramebufferTexture2D = (target, attachment, textarget, texture, level) => {
	GLctx.framebufferTexture2D(target, attachment, textarget, GL.textures[texture], level);
};
var _emscripten_glFramebufferTexture2D = _glFramebufferTexture2D;
function _glFrontFace(x0) {
	GLctx.frontFace(x0);
}
var _emscripten_glFrontFace = _glFrontFace;
var __glGenObject = (n, buffers, createFunction, objectTable) => {
	for (var i = 0; i < n; i++) {
		var buffer = GLctx[createFunction]();
		var id = buffer && GL.getNewId(objectTable);
		if (buffer) {
			buffer.name = id;
			objectTable[id] = buffer;
		} else {
			GL.recordError(1282);
		}
		HEAP32[(buffers + i * 4) >> 2] = id;
	}
};
var _glGenBuffers = (n, buffers) => {
	__glGenObject(n, buffers, 'createBuffer', GL.buffers);
};
var _emscripten_glGenBuffers = _glGenBuffers;
var _glGenFramebuffers = (n, ids) => {
	__glGenObject(n, ids, 'createFramebuffer', GL.framebuffers);
};
var _emscripten_glGenFramebuffers = _glGenFramebuffers;
var _glGenQueriesEXT = (n, ids) => {
	for (var i = 0; i < n; i++) {
		var query = GLctx.disjointTimerQueryExt['createQueryEXT']();
		if (!query) {
			GL.recordError(1282);
			while (i < n) HEAP32[(ids + i++ * 4) >> 2] = 0;
			return;
		}
		var id = GL.getNewId(GL.queries);
		query.name = id;
		GL.queries[id] = query;
		HEAP32[(ids + i * 4) >> 2] = id;
	}
};
var _emscripten_glGenQueriesEXT = _glGenQueriesEXT;
var _glGenRenderbuffers = (n, renderbuffers) => {
	__glGenObject(n, renderbuffers, 'createRenderbuffer', GL.renderbuffers);
};
var _emscripten_glGenRenderbuffers = _glGenRenderbuffers;
var _glGenTextures = (n, textures) => {
	__glGenObject(n, textures, 'createTexture', GL.textures);
};
var _emscripten_glGenTextures = _glGenTextures;
function _glGenVertexArrays(n, arrays) {
	__glGenObject(n, arrays, 'createVertexArray', GL.vaos);
}
var _glGenVertexArraysOES = _glGenVertexArrays;
var _emscripten_glGenVertexArraysOES = _glGenVertexArraysOES;
function _glGenerateMipmap(x0) {
	GLctx.generateMipmap(x0);
}
var _emscripten_glGenerateMipmap = _glGenerateMipmap;
var __glGetActiveAttribOrUniform = (
	funcName,
	program,
	index,
	bufSize,
	length,
	size,
	type,
	name
) => {
	program = GL.programs[program];
	var info = GLctx[funcName](program, index);
	if (info) {
		var numBytesWrittenExclNull = name && stringToUTF8(info.name, name, bufSize);
		if (length) HEAP32[length >> 2] = numBytesWrittenExclNull;
		if (size) HEAP32[size >> 2] = info.size;
		if (type) HEAP32[type >> 2] = info.type;
	}
};
var _glGetActiveAttrib = (program, index, bufSize, length, size, type, name) => {
	__glGetActiveAttribOrUniform(
		'getActiveAttrib',
		program,
		index,
		bufSize,
		length,
		size,
		type,
		name
	);
};
var _emscripten_glGetActiveAttrib = _glGetActiveAttrib;
var _glGetActiveUniform = (program, index, bufSize, length, size, type, name) => {
	__glGetActiveAttribOrUniform(
		'getActiveUniform',
		program,
		index,
		bufSize,
		length,
		size,
		type,
		name
	);
};
var _emscripten_glGetActiveUniform = _glGetActiveUniform;
var _glGetAttachedShaders = (program, maxCount, count, shaders) => {
	var result = GLctx.getAttachedShaders(GL.programs[program]);
	var len = result.length;
	if (len > maxCount) {
		len = maxCount;
	}
	HEAP32[count >> 2] = len;
	for (var i = 0; i < len; ++i) {
		var id = GL.shaders.indexOf(result[i]);
		HEAP32[(shaders + i * 4) >> 2] = id;
	}
};
var _emscripten_glGetAttachedShaders = _glGetAttachedShaders;
var _glGetAttribLocation = (program, name) =>
	GLctx.getAttribLocation(GL.programs[program], UTF8ToString(name));
var _emscripten_glGetAttribLocation = _glGetAttribLocation;
var readI53FromI64 = (ptr) => HEAPU32[ptr >> 2] + HEAP32[(ptr + 4) >> 2] * 4294967296;
var readI53FromU64 = (ptr) => HEAPU32[ptr >> 2] + HEAPU32[(ptr + 4) >> 2] * 4294967296;
var writeI53ToI64 = (ptr, num) => {
	HEAPU32[ptr >> 2] = num;
	var lower = HEAPU32[ptr >> 2];
	HEAPU32[(ptr + 4) >> 2] = (num - lower) / 4294967296;
	var deserialized = num >= 0 ? readI53FromU64(ptr) : readI53FromI64(ptr);
	var offset = ptr >> 2;
	if (deserialized != num)
		warnOnce(
			`writeI53ToI64() out of range: serialized JS Number ${num} to Wasm heap as bytes lo=${ptrToString(
				HEAPU32[offset]
			)}, hi=${ptrToString(
				HEAPU32[offset + 1]
			)}, which deserializes back to ${deserialized} instead!`
		);
};
var emscriptenWebGLGet = (name_, p, type) => {
	if (!p) {
		GL.recordError(1281);
		return;
	}
	var ret = undefined;
	switch (name_) {
		case 36346:
			ret = 1;
			break;
		case 36344:
			if (type != 0 && type != 1) {
				GL.recordError(1280);
			}
			return;
		case 36345:
			ret = 0;
			break;
		case 34466:
			var formats = GLctx.getParameter(34467);
			ret = formats ? formats.length : 0;
			break;
	}
	if (ret === undefined) {
		var result = GLctx.getParameter(name_);
		switch (typeof result) {
			case 'number':
				ret = result;
				break;
			case 'boolean':
				ret = result ? 1 : 0;
				break;
			case 'string':
				GL.recordError(1280);
				return;
			case 'object':
				if (result === null) {
					switch (name_) {
						case 34964:
						case 35725:
						case 34965:
						case 36006:
						case 36007:
						case 32873:
						case 34229:
						case 34068: {
							ret = 0;
							break;
						}
						default: {
							GL.recordError(1280);
							return;
						}
					}
				} else if (
					result instanceof Float32Array ||
					result instanceof Uint32Array ||
					result instanceof Int32Array ||
					result instanceof Array
				) {
					for (var i = 0; i < result.length; ++i) {
						switch (type) {
							case 0:
								HEAP32[(p + i * 4) >> 2] = result[i];
								break;
							case 2:
								HEAPF32[(p + i * 4) >> 2] = result[i];
								break;
							case 4:
								HEAP8[(p + i) >> 0] = result[i] ? 1 : 0;
								break;
						}
					}
					return;
				} else {
					try {
						ret = result.name | 0;
					} catch (e) {
						GL.recordError(1280);
						err(
							`GL_INVALID_ENUM in glGet${type}v: Unknown object returned from WebGL getParameter(${name_})! (error: ${e})`
						);
						return;
					}
				}
				break;
			default:
				GL.recordError(1280);
				err(
					`GL_INVALID_ENUM in glGet${type}v: Native code calling glGet${type}v(${name_}) and it returns ${result} of type ${typeof result}!`
				);
				return;
		}
	}
	switch (type) {
		case 1:
			writeI53ToI64(p, ret);
			break;
		case 0:
			HEAP32[p >> 2] = ret;
			break;
		case 2:
			HEAPF32[p >> 2] = ret;
			break;
		case 4:
			HEAP8[p >> 0] = ret ? 1 : 0;
			break;
	}
};
var _glGetBooleanv = (name_, p) => emscriptenWebGLGet(name_, p, 4);
var _emscripten_glGetBooleanv = _glGetBooleanv;
var _glGetBufferParameteriv = (target, value, data) => {
	if (!data) {
		GL.recordError(1281);
		return;
	}
	HEAP32[data >> 2] = GLctx.getBufferParameter(target, value);
};
var _emscripten_glGetBufferParameteriv = _glGetBufferParameteriv;
var _glGetError = () => {
	var error = GLctx.getError() || GL.lastError;
	GL.lastError = 0;
	return error;
};
var _emscripten_glGetError = _glGetError;
var _glGetFloatv = (name_, p) => emscriptenWebGLGet(name_, p, 2);
var _emscripten_glGetFloatv = _glGetFloatv;
var _glGetFramebufferAttachmentParameteriv = (target, attachment, pname, params) => {
	var result = GLctx.getFramebufferAttachmentParameter(target, attachment, pname);
	if (result instanceof WebGLRenderbuffer || result instanceof WebGLTexture) {
		result = result.name | 0;
	}
	HEAP32[params >> 2] = result;
};
var _emscripten_glGetFramebufferAttachmentParameteriv = _glGetFramebufferAttachmentParameteriv;
var _glGetIntegerv = (name_, p) => emscriptenWebGLGet(name_, p, 0);
var _emscripten_glGetIntegerv = _glGetIntegerv;
var _glGetProgramInfoLog = (program, maxLength, length, infoLog) => {
	var log = GLctx.getProgramInfoLog(GL.programs[program]);
	if (log === null) log = '(unknown error)';
	var numBytesWrittenExclNull =
		maxLength > 0 && infoLog ? stringToUTF8(log, infoLog, maxLength) : 0;
	if (length) HEAP32[length >> 2] = numBytesWrittenExclNull;
};
var _emscripten_glGetProgramInfoLog = _glGetProgramInfoLog;
var _glGetProgramiv = (program, pname, p) => {
	if (!p) {
		GL.recordError(1281);
		return;
	}
	if (program >= GL.counter) {
		GL.recordError(1281);
		return;
	}
	program = GL.programs[program];
	if (pname == 35716) {
		var log = GLctx.getProgramInfoLog(program);
		if (log === null) log = '(unknown error)';
		HEAP32[p >> 2] = log.length + 1;
	} else if (pname == 35719) {
		if (!program.maxUniformLength) {
			for (var i = 0; i < GLctx.getProgramParameter(program, 35718); ++i) {
				program.maxUniformLength = Math.max(
					program.maxUniformLength,
					GLctx.getActiveUniform(program, i).name.length + 1
				);
			}
		}
		HEAP32[p >> 2] = program.maxUniformLength;
	} else if (pname == 35722) {
		if (!program.maxAttributeLength) {
			for (var i = 0; i < GLctx.getProgramParameter(program, 35721); ++i) {
				program.maxAttributeLength = Math.max(
					program.maxAttributeLength,
					GLctx.getActiveAttrib(program, i).name.length + 1
				);
			}
		}
		HEAP32[p >> 2] = program.maxAttributeLength;
	} else if (pname == 35381) {
		if (!program.maxUniformBlockNameLength) {
			for (var i = 0; i < GLctx.getProgramParameter(program, 35382); ++i) {
				program.maxUniformBlockNameLength = Math.max(
					program.maxUniformBlockNameLength,
					GLctx.getActiveUniformBlockName(program, i).length + 1
				);
			}
		}
		HEAP32[p >> 2] = program.maxUniformBlockNameLength;
	} else {
		HEAP32[p >> 2] = GLctx.getProgramParameter(program, pname);
	}
};
var _emscripten_glGetProgramiv = _glGetProgramiv;
var _glGetQueryObjecti64vEXT = (id, pname, params) => {
	if (!params) {
		GL.recordError(1281);
		return;
	}
	var query = GL.queries[id];
	var param;
	{
		param = GLctx.disjointTimerQueryExt['getQueryObjectEXT'](query, pname);
	}
	var ret;
	if (typeof param == 'boolean') {
		ret = param ? 1 : 0;
	} else {
		ret = param;
	}
	writeI53ToI64(params, ret);
};
var _emscripten_glGetQueryObjecti64vEXT = _glGetQueryObjecti64vEXT;
var _glGetQueryObjectivEXT = (id, pname, params) => {
	if (!params) {
		GL.recordError(1281);
		return;
	}
	var query = GL.queries[id];
	var param = GLctx.disjointTimerQueryExt['getQueryObjectEXT'](query, pname);
	var ret;
	if (typeof param == 'boolean') {
		ret = param ? 1 : 0;
	} else {
		ret = param;
	}
	HEAP32[params >> 2] = ret;
};
var _emscripten_glGetQueryObjectivEXT = _glGetQueryObjectivEXT;
var _glGetQueryObjectui64vEXT = _glGetQueryObjecti64vEXT;
var _emscripten_glGetQueryObjectui64vEXT = _glGetQueryObjectui64vEXT;
var _glGetQueryObjectuivEXT = _glGetQueryObjectivEXT;
var _emscripten_glGetQueryObjectuivEXT = _glGetQueryObjectuivEXT;
var _glGetQueryivEXT = (target, pname, params) => {
	if (!params) {
		GL.recordError(1281);
		return;
	}
	HEAP32[params >> 2] = GLctx.disjointTimerQueryExt['getQueryEXT'](target, pname);
};
var _emscripten_glGetQueryivEXT = _glGetQueryivEXT;
var _glGetRenderbufferParameteriv = (target, pname, params) => {
	if (!params) {
		GL.recordError(1281);
		return;
	}
	HEAP32[params >> 2] = GLctx.getRenderbufferParameter(target, pname);
};
var _emscripten_glGetRenderbufferParameteriv = _glGetRenderbufferParameteriv;
var _glGetShaderInfoLog = (shader, maxLength, length, infoLog) => {
	var log = GLctx.getShaderInfoLog(GL.shaders[shader]);
	if (log === null) log = '(unknown error)';
	var numBytesWrittenExclNull =
		maxLength > 0 && infoLog ? stringToUTF8(log, infoLog, maxLength) : 0;
	if (length) HEAP32[length >> 2] = numBytesWrittenExclNull;
};
var _emscripten_glGetShaderInfoLog = _glGetShaderInfoLog;
var _glGetShaderPrecisionFormat = (shaderType, precisionType, range, precision) => {
	var result = GLctx.getShaderPrecisionFormat(shaderType, precisionType);
	HEAP32[range >> 2] = result.rangeMin;
	HEAP32[(range + 4) >> 2] = result.rangeMax;
	HEAP32[precision >> 2] = result.precision;
};
var _emscripten_glGetShaderPrecisionFormat = _glGetShaderPrecisionFormat;
var _glGetShaderSource = (shader, bufSize, length, source) => {
	var result = GLctx.getShaderSource(GL.shaders[shader]);
	if (!result) return;
	var numBytesWrittenExclNull = bufSize > 0 && source ? stringToUTF8(result, source, bufSize) : 0;
	if (length) HEAP32[length >> 2] = numBytesWrittenExclNull;
};
var _emscripten_glGetShaderSource = _glGetShaderSource;
var _glGetShaderiv = (shader, pname, p) => {
	if (!p) {
		GL.recordError(1281);
		return;
	}
	if (pname == 35716) {
		var log = GLctx.getShaderInfoLog(GL.shaders[shader]);
		if (log === null) log = '(unknown error)';
		var logLength = log ? log.length + 1 : 0;
		HEAP32[p >> 2] = logLength;
	} else if (pname == 35720) {
		var source = GLctx.getShaderSource(GL.shaders[shader]);
		var sourceLength = source ? source.length + 1 : 0;
		HEAP32[p >> 2] = sourceLength;
	} else {
		HEAP32[p >> 2] = GLctx.getShaderParameter(GL.shaders[shader], pname);
	}
};
var _emscripten_glGetShaderiv = _glGetShaderiv;
var stringToNewUTF8 = (str) => {
	var size = lengthBytesUTF8(str) + 1;
	var ret = _malloc(size);
	if (ret) stringToUTF8(str, ret, size);
	return ret;
};
var _glGetString = (name_) => {
	var ret = GL.stringCache[name_];
	if (!ret) {
		switch (name_) {
			case 7939:
				ret = stringToNewUTF8(GL.getExtensions().join(' '));
				break;
			case 7936:
			case 7937:
			case 37445:
			case 37446:
				var s = GLctx.getParameter(name_);
				if (!s) {
					GL.recordError(1280);
				}
				ret = s ? stringToNewUTF8(s) : 0;
				break;
			case 7938:
				var glVersion = GLctx.getParameter(7938);
				{
					glVersion = `OpenGL ES 2.0 (${glVersion})`;
				}
				ret = stringToNewUTF8(glVersion);
				break;
			case 35724:
				var glslVersion = GLctx.getParameter(35724);
				var ver_re = /^WebGL GLSL ES ([0-9]\.[0-9][0-9]?)(?:$| .*)/;
				var ver_num = glslVersion.match(ver_re);
				if (ver_num !== null) {
					if (ver_num[1].length == 3) ver_num[1] = ver_num[1] + '0';
					glslVersion = `OpenGL ES GLSL ES ${ver_num[1]} (${glslVersion})`;
				}
				ret = stringToNewUTF8(glslVersion);
				break;
			default:
				GL.recordError(1280);
		}
		GL.stringCache[name_] = ret;
	}
	return ret;
};
var _emscripten_glGetString = _glGetString;
var _glGetTexParameterfv = (target, pname, params) => {
	if (!params) {
		GL.recordError(1281);
		return;
	}
	HEAPF32[params >> 2] = GLctx.getTexParameter(target, pname);
};
var _emscripten_glGetTexParameterfv = _glGetTexParameterfv;
var _glGetTexParameteriv = (target, pname, params) => {
	if (!params) {
		GL.recordError(1281);
		return;
	}
	HEAP32[params >> 2] = GLctx.getTexParameter(target, pname);
};
var _emscripten_glGetTexParameteriv = _glGetTexParameteriv;
var jstoi_q = (str) => parseInt(str);
var webglGetLeftBracePos = (name) => name.slice(-1) == ']' && name.lastIndexOf('[');
var webglPrepareUniformLocationsBeforeFirstUse = (program) => {
	var uniformLocsById = program.uniformLocsById,
		uniformSizeAndIdsByName = program.uniformSizeAndIdsByName,
		i,
		j;
	if (!uniformLocsById) {
		program.uniformLocsById = uniformLocsById = {};
		program.uniformArrayNamesById = {};
		for (i = 0; i < GLctx.getProgramParameter(program, 35718); ++i) {
			var u = GLctx.getActiveUniform(program, i);
			var nm = u.name;
			var sz = u.size;
			var lb = webglGetLeftBracePos(nm);
			var arrayName = lb > 0 ? nm.slice(0, lb) : nm;
			var id = program.uniformIdCounter;
			program.uniformIdCounter += sz;
			uniformSizeAndIdsByName[arrayName] = [sz, id];
			for (j = 0; j < sz; ++j) {
				uniformLocsById[id] = j;
				program.uniformArrayNamesById[id++] = arrayName;
			}
		}
	}
};
var _glGetUniformLocation = (program, name) => {
	name = UTF8ToString(name);
	if ((program = GL.programs[program])) {
		webglPrepareUniformLocationsBeforeFirstUse(program);
		var uniformLocsById = program.uniformLocsById;
		var arrayIndex = 0;
		var uniformBaseName = name;
		var leftBrace = webglGetLeftBracePos(name);
		if (leftBrace > 0) {
			arrayIndex = jstoi_q(name.slice(leftBrace + 1)) >>> 0;
			uniformBaseName = name.slice(0, leftBrace);
		}
		var sizeAndId = program.uniformSizeAndIdsByName[uniformBaseName];
		if (sizeAndId && arrayIndex < sizeAndId[0]) {
			arrayIndex += sizeAndId[1];
			if (
				(uniformLocsById[arrayIndex] =
					uniformLocsById[arrayIndex] || GLctx.getUniformLocation(program, name))
			) {
				return arrayIndex;
			}
		}
	} else {
		GL.recordError(1281);
	}
	return -1;
};
var _emscripten_glGetUniformLocation = _glGetUniformLocation;
var webglGetUniformLocation = (location) => {
	var p = GLctx.currentProgram;
	if (p) {
		var webglLoc = p.uniformLocsById[location];
		if (typeof webglLoc == 'number') {
			p.uniformLocsById[location] = webglLoc = GLctx.getUniformLocation(
				p,
				p.uniformArrayNamesById[location] + (webglLoc > 0 ? `[${webglLoc}]` : '')
			);
		}
		return webglLoc;
	} else {
		GL.recordError(1282);
	}
};
var emscriptenWebGLGetUniform = (program, location, params, type) => {
	if (!params) {
		GL.recordError(1281);
		return;
	}
	program = GL.programs[program];
	webglPrepareUniformLocationsBeforeFirstUse(program);
	var data = GLctx.getUniform(program, webglGetUniformLocation(location));
	if (typeof data == 'number' || typeof data == 'boolean') {
		switch (type) {
			case 0:
				HEAP32[params >> 2] = data;
				break;
			case 2:
				HEAPF32[params >> 2] = data;
				break;
		}
	} else {
		for (var i = 0; i < data.length; i++) {
			switch (type) {
				case 0:
					HEAP32[(params + i * 4) >> 2] = data[i];
					break;
				case 2:
					HEAPF32[(params + i * 4) >> 2] = data[i];
					break;
			}
		}
	}
};
var _glGetUniformfv = (program, location, params) => {
	emscriptenWebGLGetUniform(program, location, params, 2);
};
var _emscripten_glGetUniformfv = _glGetUniformfv;
var _glGetUniformiv = (program, location, params) => {
	emscriptenWebGLGetUniform(program, location, params, 0);
};
var _emscripten_glGetUniformiv = _glGetUniformiv;
var _glGetVertexAttribPointerv = (index, pname, pointer) => {
	if (!pointer) {
		GL.recordError(1281);
		return;
	}
	HEAP32[pointer >> 2] = GLctx.getVertexAttribOffset(index, pname);
};
var _emscripten_glGetVertexAttribPointerv = _glGetVertexAttribPointerv;
var emscriptenWebGLGetVertexAttrib = (index, pname, params, type) => {
	if (!params) {
		GL.recordError(1281);
		return;
	}
	var data = GLctx.getVertexAttrib(index, pname);
	if (pname == 34975) {
		HEAP32[params >> 2] = data && data['name'];
	} else if (typeof data == 'number' || typeof data == 'boolean') {
		switch (type) {
			case 0:
				HEAP32[params >> 2] = data;
				break;
			case 2:
				HEAPF32[params >> 2] = data;
				break;
			case 5:
				HEAP32[params >> 2] = Math.fround(data);
				break;
		}
	} else {
		for (var i = 0; i < data.length; i++) {
			switch (type) {
				case 0:
					HEAP32[(params + i * 4) >> 2] = data[i];
					break;
				case 2:
					HEAPF32[(params + i * 4) >> 2] = data[i];
					break;
				case 5:
					HEAP32[(params + i * 4) >> 2] = Math.fround(data[i]);
					break;
			}
		}
	}
};
var _glGetVertexAttribfv = (index, pname, params) => {
	emscriptenWebGLGetVertexAttrib(index, pname, params, 2);
};
var _emscripten_glGetVertexAttribfv = _glGetVertexAttribfv;
var _glGetVertexAttribiv = (index, pname, params) => {
	emscriptenWebGLGetVertexAttrib(index, pname, params, 5);
};
var _emscripten_glGetVertexAttribiv = _glGetVertexAttribiv;
function _glHint(x0, x1) {
	GLctx.hint(x0, x1);
}
var _emscripten_glHint = _glHint;
var _glIsBuffer = (buffer) => {
	var b = GL.buffers[buffer];
	if (!b) return 0;
	return GLctx.isBuffer(b);
};
var _emscripten_glIsBuffer = _glIsBuffer;
function _glIsEnabled(x0) {
	return GLctx.isEnabled(x0);
}
var _emscripten_glIsEnabled = _glIsEnabled;
var _glIsFramebuffer = (framebuffer) => {
	var fb = GL.framebuffers[framebuffer];
	if (!fb) return 0;
	return GLctx.isFramebuffer(fb);
};
var _emscripten_glIsFramebuffer = _glIsFramebuffer;
var _glIsProgram = (program) => {
	program = GL.programs[program];
	if (!program) return 0;
	return GLctx.isProgram(program);
};
var _emscripten_glIsProgram = _glIsProgram;
var _glIsQueryEXT = (id) => {
	var query = GL.queries[id];
	if (!query) return 0;
	return GLctx.disjointTimerQueryExt['isQueryEXT'](query);
};
var _emscripten_glIsQueryEXT = _glIsQueryEXT;
var _glIsRenderbuffer = (renderbuffer) => {
	var rb = GL.renderbuffers[renderbuffer];
	if (!rb) return 0;
	return GLctx.isRenderbuffer(rb);
};
var _emscripten_glIsRenderbuffer = _glIsRenderbuffer;
var _glIsShader = (shader) => {
	var s = GL.shaders[shader];
	if (!s) return 0;
	return GLctx.isShader(s);
};
var _emscripten_glIsShader = _glIsShader;
var _glIsTexture = (id) => {
	var texture = GL.textures[id];
	if (!texture) return 0;
	return GLctx.isTexture(texture);
};
var _emscripten_glIsTexture = _glIsTexture;
var _glIsVertexArray = (array) => {
	var vao = GL.vaos[array];
	if (!vao) return 0;
	return GLctx.isVertexArray(vao);
};
var _glIsVertexArrayOES = _glIsVertexArray;
var _emscripten_glIsVertexArrayOES = _glIsVertexArrayOES;
function _glLineWidth(x0) {
	GLctx.lineWidth(x0);
}
var _emscripten_glLineWidth = _glLineWidth;
var _glLinkProgram = (program) => {
	program = GL.programs[program];
	GLctx.linkProgram(program);
	program.uniformLocsById = 0;
	program.uniformSizeAndIdsByName = {};
};
var _emscripten_glLinkProgram = _glLinkProgram;
var _glPixelStorei = (pname, param) => {
	if (pname == 3317) {
		GL.unpackAlignment = param;
	}
	GLctx.pixelStorei(pname, param);
};
var _emscripten_glPixelStorei = _glPixelStorei;
function _glPolygonOffset(x0, x1) {
	GLctx.polygonOffset(x0, x1);
}
var _emscripten_glPolygonOffset = _glPolygonOffset;
var _glQueryCounterEXT = (id, target) => {
	GLctx.disjointTimerQueryExt['queryCounterEXT'](GL.queries[id], target);
};
var _emscripten_glQueryCounterEXT = _glQueryCounterEXT;
var computeUnpackAlignedImageSize = (width, height, sizePerPixel, alignment) => {
	function roundedToNextMultipleOf(x, y) {
		return (x + y - 1) & -y;
	}
	var plainRowSize = width * sizePerPixel;
	var alignedRowSize = roundedToNextMultipleOf(plainRowSize, alignment);
	return height * alignedRowSize;
};
var colorChannelsInGlTextureFormat = (format) => {
	var colorChannels = { 5: 3, 6: 4, 8: 2, 29502: 3, 29504: 4 };
	return colorChannels[format - 6402] || 1;
};
var heapObjectForWebGLType = (type) => {
	type -= 5120;
	if (type == 1) return HEAPU8;
	if (type == 4) return HEAP32;
	if (type == 6) return HEAPF32;
	if (type == 5 || type == 28922) return HEAPU32;
	return HEAPU16;
};
var heapAccessShiftForWebGLHeap = (heap) => 31 - Math.clz32(heap.BYTES_PER_ELEMENT);
var emscriptenWebGLGetTexPixelData = (type, format, width, height, pixels, internalFormat) => {
	var heap = heapObjectForWebGLType(type);
	var shift = heapAccessShiftForWebGLHeap(heap);
	var byteSize = 1 << shift;
	var sizePerPixel = colorChannelsInGlTextureFormat(format) * byteSize;
	var bytes = computeUnpackAlignedImageSize(width, height, sizePerPixel, GL.unpackAlignment);
	return heap.subarray(pixels >> shift, (pixels + bytes) >> shift);
};
var _glReadPixels = (x, y, width, height, format, type, pixels) => {
	var pixelData = emscriptenWebGLGetTexPixelData(type, format, width, height, pixels, format);
	if (!pixelData) {
		GL.recordError(1280);
		return;
	}
	GLctx.readPixels(x, y, width, height, format, type, pixelData);
};
var _emscripten_glReadPixels = _glReadPixels;
var _glReleaseShaderCompiler = () => {};
var _emscripten_glReleaseShaderCompiler = _glReleaseShaderCompiler;
function _glRenderbufferStorage(x0, x1, x2, x3) {
	GLctx.renderbufferStorage(x0, x1, x2, x3);
}
var _emscripten_glRenderbufferStorage = _glRenderbufferStorage;
var _glSampleCoverage = (value, invert) => {
	GLctx.sampleCoverage(value, !!invert);
};
var _emscripten_glSampleCoverage = _glSampleCoverage;
function _glScissor(x0, x1, x2, x3) {
	GLctx.scissor(x0, x1, x2, x3);
}
var _emscripten_glScissor = _glScissor;
var _glShaderBinary = (count, shaders, binaryformat, binary, length) => {
	GL.recordError(1280);
};
var _emscripten_glShaderBinary = _glShaderBinary;
var _glShaderSource = (shader, count, string, length) => {
	var source = GL.getSource(shader, count, string, length);
	GLctx.shaderSource(GL.shaders[shader], source);
};
var _emscripten_glShaderSource = _glShaderSource;
function _glStencilFunc(x0, x1, x2) {
	GLctx.stencilFunc(x0, x1, x2);
}
var _emscripten_glStencilFunc = _glStencilFunc;
function _glStencilFuncSeparate(x0, x1, x2, x3) {
	GLctx.stencilFuncSeparate(x0, x1, x2, x3);
}
var _emscripten_glStencilFuncSeparate = _glStencilFuncSeparate;
function _glStencilMask(x0) {
	GLctx.stencilMask(x0);
}
var _emscripten_glStencilMask = _glStencilMask;
function _glStencilMaskSeparate(x0, x1) {
	GLctx.stencilMaskSeparate(x0, x1);
}
var _emscripten_glStencilMaskSeparate = _glStencilMaskSeparate;
function _glStencilOp(x0, x1, x2) {
	GLctx.stencilOp(x0, x1, x2);
}
var _emscripten_glStencilOp = _glStencilOp;
function _glStencilOpSeparate(x0, x1, x2, x3) {
	GLctx.stencilOpSeparate(x0, x1, x2, x3);
}
var _emscripten_glStencilOpSeparate = _glStencilOpSeparate;
var _glTexImage2D = (
	target,
	level,
	internalFormat,
	width,
	height,
	border,
	format,
	type,
	pixels
) => {
	GLctx.texImage2D(
		target,
		level,
		internalFormat,
		width,
		height,
		border,
		format,
		type,
		pixels
			? emscriptenWebGLGetTexPixelData(type, format, width, height, pixels, internalFormat)
			: null
	);
};
var _emscripten_glTexImage2D = _glTexImage2D;
function _glTexParameterf(x0, x1, x2) {
	GLctx.texParameterf(x0, x1, x2);
}
var _emscripten_glTexParameterf = _glTexParameterf;
var _glTexParameterfv = (target, pname, params) => {
	var param = HEAPF32[params >> 2];
	GLctx.texParameterf(target, pname, param);
};
var _emscripten_glTexParameterfv = _glTexParameterfv;
function _glTexParameteri(x0, x1, x2) {
	GLctx.texParameteri(x0, x1, x2);
}
var _emscripten_glTexParameteri = _glTexParameteri;
var _glTexParameteriv = (target, pname, params) => {
	var param = HEAP32[params >> 2];
	GLctx.texParameteri(target, pname, param);
};
var _emscripten_glTexParameteriv = _glTexParameteriv;
var _glTexSubImage2D = (target, level, xoffset, yoffset, width, height, format, type, pixels) => {
	var pixelData = null;
	if (pixels) pixelData = emscriptenWebGLGetTexPixelData(type, format, width, height, pixels, 0);
	GLctx.texSubImage2D(target, level, xoffset, yoffset, width, height, format, type, pixelData);
};
var _emscripten_glTexSubImage2D = _glTexSubImage2D;
var _glUniform1f = (location, v0) => {
	GLctx.uniform1f(webglGetUniformLocation(location), v0);
};
var _emscripten_glUniform1f = _glUniform1f;
var miniTempWebGLFloatBuffers = [];
var _glUniform1fv = (location, count, value) => {
	if (count <= 288) {
		var view = miniTempWebGLFloatBuffers[count - 1];
		for (var i = 0; i < count; ++i) {
			view[i] = HEAPF32[(value + 4 * i) >> 2];
		}
	} else {
		var view = HEAPF32.subarray(value >> 2, (value + count * 4) >> 2);
	}
	GLctx.uniform1fv(webglGetUniformLocation(location), view);
};
var _emscripten_glUniform1fv = _glUniform1fv;
var _glUniform1i = (location, v0) => {
	GLctx.uniform1i(webglGetUniformLocation(location), v0);
};
var _emscripten_glUniform1i = _glUniform1i;
var miniTempWebGLIntBuffers = [];
var _glUniform1iv = (location, count, value) => {
	if (count <= 288) {
		var view = miniTempWebGLIntBuffers[count - 1];
		for (var i = 0; i < count; ++i) {
			view[i] = HEAP32[(value + 4 * i) >> 2];
		}
	} else {
		var view = HEAP32.subarray(value >> 2, (value + count * 4) >> 2);
	}
	GLctx.uniform1iv(webglGetUniformLocation(location), view);
};
var _emscripten_glUniform1iv = _glUniform1iv;
var _glUniform2f = (location, v0, v1) => {
	GLctx.uniform2f(webglGetUniformLocation(location), v0, v1);
};
var _emscripten_glUniform2f = _glUniform2f;
var _glUniform2fv = (location, count, value) => {
	if (count <= 144) {
		var view = miniTempWebGLFloatBuffers[2 * count - 1];
		for (var i = 0; i < 2 * count; i += 2) {
			view[i] = HEAPF32[(value + 4 * i) >> 2];
			view[i + 1] = HEAPF32[(value + (4 * i + 4)) >> 2];
		}
	} else {
		var view = HEAPF32.subarray(value >> 2, (value + count * 8) >> 2);
	}
	GLctx.uniform2fv(webglGetUniformLocation(location), view);
};
var _emscripten_glUniform2fv = _glUniform2fv;
var _glUniform2i = (location, v0, v1) => {
	GLctx.uniform2i(webglGetUniformLocation(location), v0, v1);
};
var _emscripten_glUniform2i = _glUniform2i;
var _glUniform2iv = (location, count, value) => {
	if (count <= 144) {
		var view = miniTempWebGLIntBuffers[2 * count - 1];
		for (var i = 0; i < 2 * count; i += 2) {
			view[i] = HEAP32[(value + 4 * i) >> 2];
			view[i + 1] = HEAP32[(value + (4 * i + 4)) >> 2];
		}
	} else {
		var view = HEAP32.subarray(value >> 2, (value + count * 8) >> 2);
	}
	GLctx.uniform2iv(webglGetUniformLocation(location), view);
};
var _emscripten_glUniform2iv = _glUniform2iv;
var _glUniform3f = (location, v0, v1, v2) => {
	GLctx.uniform3f(webglGetUniformLocation(location), v0, v1, v2);
};
var _emscripten_glUniform3f = _glUniform3f;
var _glUniform3fv = (location, count, value) => {
	if (count <= 96) {
		var view = miniTempWebGLFloatBuffers[3 * count - 1];
		for (var i = 0; i < 3 * count; i += 3) {
			view[i] = HEAPF32[(value + 4 * i) >> 2];
			view[i + 1] = HEAPF32[(value + (4 * i + 4)) >> 2];
			view[i + 2] = HEAPF32[(value + (4 * i + 8)) >> 2];
		}
	} else {
		var view = HEAPF32.subarray(value >> 2, (value + count * 12) >> 2);
	}
	GLctx.uniform3fv(webglGetUniformLocation(location), view);
};
var _emscripten_glUniform3fv = _glUniform3fv;
var _glUniform3i = (location, v0, v1, v2) => {
	GLctx.uniform3i(webglGetUniformLocation(location), v0, v1, v2);
};
var _emscripten_glUniform3i = _glUniform3i;
var _glUniform3iv = (location, count, value) => {
	if (count <= 96) {
		var view = miniTempWebGLIntBuffers[3 * count - 1];
		for (var i = 0; i < 3 * count; i += 3) {
			view[i] = HEAP32[(value + 4 * i) >> 2];
			view[i + 1] = HEAP32[(value + (4 * i + 4)) >> 2];
			view[i + 2] = HEAP32[(value + (4 * i + 8)) >> 2];
		}
	} else {
		var view = HEAP32.subarray(value >> 2, (value + count * 12) >> 2);
	}
	GLctx.uniform3iv(webglGetUniformLocation(location), view);
};
var _emscripten_glUniform3iv = _glUniform3iv;
var _glUniform4f = (location, v0, v1, v2, v3) => {
	GLctx.uniform4f(webglGetUniformLocation(location), v0, v1, v2, v3);
};
var _emscripten_glUniform4f = _glUniform4f;
var _glUniform4fv = (location, count, value) => {
	if (count <= 72) {
		var view = miniTempWebGLFloatBuffers[4 * count - 1];
		var heap = HEAPF32;
		value >>= 2;
		for (var i = 0; i < 4 * count; i += 4) {
			var dst = value + i;
			view[i] = heap[dst];
			view[i + 1] = heap[dst + 1];
			view[i + 2] = heap[dst + 2];
			view[i + 3] = heap[dst + 3];
		}
	} else {
		var view = HEAPF32.subarray(value >> 2, (value + count * 16) >> 2);
	}
	GLctx.uniform4fv(webglGetUniformLocation(location), view);
};
var _emscripten_glUniform4fv = _glUniform4fv;
var _glUniform4i = (location, v0, v1, v2, v3) => {
	GLctx.uniform4i(webglGetUniformLocation(location), v0, v1, v2, v3);
};
var _emscripten_glUniform4i = _glUniform4i;
var _glUniform4iv = (location, count, value) => {
	if (count <= 72) {
		var view = miniTempWebGLIntBuffers[4 * count - 1];
		for (var i = 0; i < 4 * count; i += 4) {
			view[i] = HEAP32[(value + 4 * i) >> 2];
			view[i + 1] = HEAP32[(value + (4 * i + 4)) >> 2];
			view[i + 2] = HEAP32[(value + (4 * i + 8)) >> 2];
			view[i + 3] = HEAP32[(value + (4 * i + 12)) >> 2];
		}
	} else {
		var view = HEAP32.subarray(value >> 2, (value + count * 16) >> 2);
	}
	GLctx.uniform4iv(webglGetUniformLocation(location), view);
};
var _emscripten_glUniform4iv = _glUniform4iv;
var _glUniformMatrix2fv = (location, count, transpose, value) => {
	if (count <= 72) {
		var view = miniTempWebGLFloatBuffers[4 * count - 1];
		for (var i = 0; i < 4 * count; i += 4) {
			view[i] = HEAPF32[(value + 4 * i) >> 2];
			view[i + 1] = HEAPF32[(value + (4 * i + 4)) >> 2];
			view[i + 2] = HEAPF32[(value + (4 * i + 8)) >> 2];
			view[i + 3] = HEAPF32[(value + (4 * i + 12)) >> 2];
		}
	} else {
		var view = HEAPF32.subarray(value >> 2, (value + count * 16) >> 2);
	}
	GLctx.uniformMatrix2fv(webglGetUniformLocation(location), !!transpose, view);
};
var _emscripten_glUniformMatrix2fv = _glUniformMatrix2fv;
var _glUniformMatrix3fv = (location, count, transpose, value) => {
	if (count <= 32) {
		var view = miniTempWebGLFloatBuffers[9 * count - 1];
		for (var i = 0; i < 9 * count; i += 9) {
			view[i] = HEAPF32[(value + 4 * i) >> 2];
			view[i + 1] = HEAPF32[(value + (4 * i + 4)) >> 2];
			view[i + 2] = HEAPF32[(value + (4 * i + 8)) >> 2];
			view[i + 3] = HEAPF32[(value + (4 * i + 12)) >> 2];
			view[i + 4] = HEAPF32[(value + (4 * i + 16)) >> 2];
			view[i + 5] = HEAPF32[(value + (4 * i + 20)) >> 2];
			view[i + 6] = HEAPF32[(value + (4 * i + 24)) >> 2];
			view[i + 7] = HEAPF32[(value + (4 * i + 28)) >> 2];
			view[i + 8] = HEAPF32[(value + (4 * i + 32)) >> 2];
		}
	} else {
		var view = HEAPF32.subarray(value >> 2, (value + count * 36) >> 2);
	}
	GLctx.uniformMatrix3fv(webglGetUniformLocation(location), !!transpose, view);
};
var _emscripten_glUniformMatrix3fv = _glUniformMatrix3fv;
var _glUniformMatrix4fv = (location, count, transpose, value) => {
	if (count <= 18) {
		var view = miniTempWebGLFloatBuffers[16 * count - 1];
		var heap = HEAPF32;
		value >>= 2;
		for (var i = 0; i < 16 * count; i += 16) {
			var dst = value + i;
			view[i] = heap[dst];
			view[i + 1] = heap[dst + 1];
			view[i + 2] = heap[dst + 2];
			view[i + 3] = heap[dst + 3];
			view[i + 4] = heap[dst + 4];
			view[i + 5] = heap[dst + 5];
			view[i + 6] = heap[dst + 6];
			view[i + 7] = heap[dst + 7];
			view[i + 8] = heap[dst + 8];
			view[i + 9] = heap[dst + 9];
			view[i + 10] = heap[dst + 10];
			view[i + 11] = heap[dst + 11];
			view[i + 12] = heap[dst + 12];
			view[i + 13] = heap[dst + 13];
			view[i + 14] = heap[dst + 14];
			view[i + 15] = heap[dst + 15];
		}
	} else {
		var view = HEAPF32.subarray(value >> 2, (value + count * 64) >> 2);
	}
	GLctx.uniformMatrix4fv(webglGetUniformLocation(location), !!transpose, view);
};
var _emscripten_glUniformMatrix4fv = _glUniformMatrix4fv;
var _glUseProgram = (program) => {
	program = GL.programs[program];
	GLctx.useProgram(program);
	GLctx.currentProgram = program;
};
var _emscripten_glUseProgram = _glUseProgram;
var _glValidateProgram = (program) => {
	GLctx.validateProgram(GL.programs[program]);
};
var _emscripten_glValidateProgram = _glValidateProgram;
function _glVertexAttrib1f(x0, x1) {
	GLctx.vertexAttrib1f(x0, x1);
}
var _emscripten_glVertexAttrib1f = _glVertexAttrib1f;
var _glVertexAttrib1fv = (index, v) => {
	GLctx.vertexAttrib1f(index, HEAPF32[v >> 2]);
};
var _emscripten_glVertexAttrib1fv = _glVertexAttrib1fv;
function _glVertexAttrib2f(x0, x1, x2) {
	GLctx.vertexAttrib2f(x0, x1, x2);
}
var _emscripten_glVertexAttrib2f = _glVertexAttrib2f;
var _glVertexAttrib2fv = (index, v) => {
	GLctx.vertexAttrib2f(index, HEAPF32[v >> 2], HEAPF32[(v + 4) >> 2]);
};
var _emscripten_glVertexAttrib2fv = _glVertexAttrib2fv;
function _glVertexAttrib3f(x0, x1, x2, x3) {
	GLctx.vertexAttrib3f(x0, x1, x2, x3);
}
var _emscripten_glVertexAttrib3f = _glVertexAttrib3f;
var _glVertexAttrib3fv = (index, v) => {
	GLctx.vertexAttrib3f(index, HEAPF32[v >> 2], HEAPF32[(v + 4) >> 2], HEAPF32[(v + 8) >> 2]);
};
var _emscripten_glVertexAttrib3fv = _glVertexAttrib3fv;
function _glVertexAttrib4f(x0, x1, x2, x3, x4) {
	GLctx.vertexAttrib4f(x0, x1, x2, x3, x4);
}
var _emscripten_glVertexAttrib4f = _glVertexAttrib4f;
var _glVertexAttrib4fv = (index, v) => {
	GLctx.vertexAttrib4f(
		index,
		HEAPF32[v >> 2],
		HEAPF32[(v + 4) >> 2],
		HEAPF32[(v + 8) >> 2],
		HEAPF32[(v + 12) >> 2]
	);
};
var _emscripten_glVertexAttrib4fv = _glVertexAttrib4fv;
var _glVertexAttribDivisor = (index, divisor) => {
	GLctx.vertexAttribDivisor(index, divisor);
};
var _glVertexAttribDivisorANGLE = _glVertexAttribDivisor;
var _emscripten_glVertexAttribDivisorANGLE = _glVertexAttribDivisorANGLE;
var _glVertexAttribPointer = (index, size, type, normalized, stride, ptr) => {
	GLctx.vertexAttribPointer(index, size, type, !!normalized, stride, ptr);
};
var _emscripten_glVertexAttribPointer = _glVertexAttribPointer;
function _glViewport(x0, x1, x2, x3) {
	GLctx.viewport(x0, x1, x2, x3);
}
var _emscripten_glViewport = _glViewport;
var _emscripten_num_logical_cores = () =>
	ENVIRONMENT_IS_NODE ? require('os').cpus().length : navigator['hardwareConcurrency'];
var abortOnCannotGrowMemory = (requestedSize) => {
	abort(
		`Cannot enlarge memory arrays to size ${requestedSize} bytes (OOM). Either (1) compile with -sINITIAL_MEMORY=X with X higher than the current value ${HEAP8.length}, (2) compile with -sALLOW_MEMORY_GROWTH which allows increasing the size at runtime, or (3) if you want malloc to return NULL (0) instead of this abort, compile with -sABORTING_MALLOC=0`
	);
};
var _emscripten_resize_heap = (requestedSize) => {
	var oldSize = HEAPU8.length;
	requestedSize >>>= 0;
	abortOnCannotGrowMemory(requestedSize);
};
var disableGamepadApiIfItThrows = () => {
	try {
		navigator.getGamepads();
	} catch (e) {
		err(
			`navigator.getGamepads() exists, but failed to execute with exception ${e}. Disabling Gamepad access.`
		);
		navigator.getGamepads = null;
		return 1;
	}
};
function _emscripten_sample_gamepad_data() {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(11, 1);
	if (!navigator.getGamepads || disableGamepadApiIfItThrows()) return -1;
	return (JSEvents.lastGamepadState = navigator.getGamepads()) ? 0 : -1;
}
var findCanvasEventTarget = (target) => findEventTarget(target);
var setCanvasElementSizeCallingThread = (target, width, height) => {
	var canvas = findCanvasEventTarget(target);
	if (!canvas) return -4;
	if (!canvas.controlTransferredOffscreen) {
		var autoResizeViewport = false;
		if (canvas.GLctxObject?.GLctx) {
			var prevViewport = canvas.GLctxObject.GLctx.getParameter(2978);
			autoResizeViewport =
				prevViewport[0] === 0 &&
				prevViewport[1] === 0 &&
				prevViewport[2] === canvas.width &&
				prevViewport[3] === canvas.height;
		}
		canvas.width = width;
		canvas.height = height;
		if (autoResizeViewport) {
			canvas.GLctxObject.GLctx.viewport(0, 0, width, height);
		}
	} else {
		return -4;
	}
	return 0;
};
function setCanvasElementSizeMainThread(target, width, height) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(12, 1, target, width, height);
	return setCanvasElementSizeCallingThread(target, width, height);
}
var _emscripten_set_canvas_element_size = (target, width, height) => {
	var canvas = findCanvasEventTarget(target);
	if (canvas) {
		return setCanvasElementSizeCallingThread(target, width, height);
	}
	return setCanvasElementSizeMainThread(target, width, height);
};
var fillMouseEventData = (eventStruct, e, target) => {
	assert(eventStruct % 4 == 0);
	HEAPF64[eventStruct >> 3] = e.timeStamp;
	var idx = eventStruct >> 2;
	HEAP32[idx + 2] = e.screenX;
	HEAP32[idx + 3] = e.screenY;
	HEAP32[idx + 4] = e.clientX;
	HEAP32[idx + 5] = e.clientY;
	HEAP32[idx + 6] = e.ctrlKey;
	HEAP32[idx + 7] = e.shiftKey;
	HEAP32[idx + 8] = e.altKey;
	HEAP32[idx + 9] = e.metaKey;
	HEAP16[idx * 2 + 20] = e.button;
	HEAP16[idx * 2 + 21] = e.buttons;
	HEAP32[idx + 11] = e['movementX'];
	HEAP32[idx + 12] = e['movementY'];
	var rect = getBoundingClientRect(target);
	HEAP32[idx + 13] = e.clientX - rect.left;
	HEAP32[idx + 14] = e.clientY - rect.top;
};
var registerMouseEventCallback = (
	target,
	userData,
	useCapture,
	callbackfunc,
	eventTypeId,
	eventTypeString,
	targetThread
) => {
	targetThread = JSEvents.getTargetThreadForEventCallback(targetThread);
	if (!JSEvents.mouseEvent) JSEvents.mouseEvent = _malloc(72);
	target = findEventTarget(target);
	var mouseEventHandlerFunc = (e = event) => {
		fillMouseEventData(JSEvents.mouseEvent, e, target);
		if (targetThread) {
			var mouseEventData = _malloc(72);
			fillMouseEventData(mouseEventData, e, target);
			__emscripten_run_callback_on_thread(
				targetThread,
				callbackfunc,
				eventTypeId,
				mouseEventData,
				userData
			);
		} else if (
			((a1, a2, a3) => dynCall_iiii.apply(null, [callbackfunc, a1, a2, a3]))(
				eventTypeId,
				JSEvents.mouseEvent,
				userData
			)
		)
			e.preventDefault();
	};
	var eventHandler = {
		target: target,
		allowsDeferredCalls:
			eventTypeString != 'mousemove' &&
			eventTypeString != 'mouseenter' &&
			eventTypeString != 'mouseleave',
		eventTypeString: eventTypeString,
		callbackfunc: callbackfunc,
		handlerFunc: mouseEventHandlerFunc,
		useCapture: useCapture
	};
	return JSEvents.registerOrRemoveHandler(eventHandler);
};
function _emscripten_set_click_callback_on_thread(
	target,
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(13, 1, target, userData, useCapture, callbackfunc, targetThread);
	return registerMouseEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		4,
		'click',
		targetThread
	);
}
var fillFullscreenChangeEventData = (eventStruct) => {
	var fullscreenElement =
		document.fullscreenElement ||
		document.mozFullScreenElement ||
		document.webkitFullscreenElement ||
		document.msFullscreenElement;
	var isFullscreen = !!fullscreenElement;
	HEAP32[eventStruct >> 2] = isFullscreen;
	HEAP32[(eventStruct + 4) >> 2] = JSEvents.fullscreenEnabled();
	var reportedElement = isFullscreen ? fullscreenElement : JSEvents.previousFullscreenElement;
	var nodeName = JSEvents.getNodeNameForTarget(reportedElement);
	var id = reportedElement?.id || '';
	stringToUTF8(nodeName, eventStruct + 8, 128);
	stringToUTF8(id, eventStruct + 136, 128);
	HEAP32[(eventStruct + 264) >> 2] = reportedElement ? reportedElement.clientWidth : 0;
	HEAP32[(eventStruct + 268) >> 2] = reportedElement ? reportedElement.clientHeight : 0;
	HEAP32[(eventStruct + 272) >> 2] = screen.width;
	HEAP32[(eventStruct + 276) >> 2] = screen.height;
	if (isFullscreen) {
		JSEvents.previousFullscreenElement = fullscreenElement;
	}
};
var registerFullscreenChangeEventCallback = (
	target,
	userData,
	useCapture,
	callbackfunc,
	eventTypeId,
	eventTypeString,
	targetThread
) => {
	targetThread = JSEvents.getTargetThreadForEventCallback(targetThread);
	if (!JSEvents.fullscreenChangeEvent) JSEvents.fullscreenChangeEvent = _malloc(280);
	var fullscreenChangeEventhandlerFunc = (e = event) => {
		var fullscreenChangeEvent = targetThread ? _malloc(280) : JSEvents.fullscreenChangeEvent;
		fillFullscreenChangeEventData(fullscreenChangeEvent);
		if (targetThread)
			__emscripten_run_callback_on_thread(
				targetThread,
				callbackfunc,
				eventTypeId,
				fullscreenChangeEvent,
				userData
			);
		else if (
			((a1, a2, a3) => dynCall_iiii.apply(null, [callbackfunc, a1, a2, a3]))(
				eventTypeId,
				fullscreenChangeEvent,
				userData
			)
		)
			e.preventDefault();
	};
	var eventHandler = {
		target: target,
		eventTypeString: eventTypeString,
		callbackfunc: callbackfunc,
		handlerFunc: fullscreenChangeEventhandlerFunc,
		useCapture: useCapture
	};
	return JSEvents.registerOrRemoveHandler(eventHandler);
};
function _emscripten_set_fullscreenchange_callback_on_thread(
	target,
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(14, 1, target, userData, useCapture, callbackfunc, targetThread);
	if (!JSEvents.fullscreenEnabled()) return -1;
	target = findEventTarget(target);
	if (!target) return -4;
	registerFullscreenChangeEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		19,
		'webkitfullscreenchange',
		targetThread
	);
	return registerFullscreenChangeEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		19,
		'fullscreenchange',
		targetThread
	);
}
var registerGamepadEventCallback = (
	target,
	userData,
	useCapture,
	callbackfunc,
	eventTypeId,
	eventTypeString,
	targetThread
) => {
	targetThread = JSEvents.getTargetThreadForEventCallback(targetThread);
	if (!JSEvents.gamepadEvent) JSEvents.gamepadEvent = _malloc(1432);
	var gamepadEventHandlerFunc = (e = event) => {
		var gamepadEvent = targetThread ? _malloc(1432) : JSEvents.gamepadEvent;
		fillGamepadEventData(gamepadEvent, e['gamepad']);
		if (targetThread)
			__emscripten_run_callback_on_thread(
				targetThread,
				callbackfunc,
				eventTypeId,
				gamepadEvent,
				userData
			);
		else if (
			((a1, a2, a3) => dynCall_iiii.apply(null, [callbackfunc, a1, a2, a3]))(
				eventTypeId,
				gamepadEvent,
				userData
			)
		)
			e.preventDefault();
	};
	var eventHandler = {
		target: findEventTarget(target),
		allowsDeferredCalls: true,
		eventTypeString: eventTypeString,
		callbackfunc: callbackfunc,
		handlerFunc: gamepadEventHandlerFunc,
		useCapture: useCapture
	};
	return JSEvents.registerOrRemoveHandler(eventHandler);
};
function _emscripten_set_gamepadconnected_callback_on_thread(
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(15, 1, userData, useCapture, callbackfunc, targetThread);
	if (!navigator.getGamepads || disableGamepadApiIfItThrows()) return -1;
	return registerGamepadEventCallback(
		2,
		userData,
		useCapture,
		callbackfunc,
		26,
		'gamepadconnected',
		targetThread
	);
}
function _emscripten_set_gamepaddisconnected_callback_on_thread(
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(16, 1, userData, useCapture, callbackfunc, targetThread);
	if (!navigator.getGamepads || disableGamepadApiIfItThrows()) return -1;
	return registerGamepadEventCallback(
		2,
		userData,
		useCapture,
		callbackfunc,
		27,
		'gamepaddisconnected',
		targetThread
	);
}
var runtimeKeepalivePop = () => {
	assert(runtimeKeepaliveCounter > 0);
	runtimeKeepaliveCounter -= 1;
};
var safeSetTimeout = (func, timeout) => {
	runtimeKeepalivePush();
	return setTimeout(() => {
		runtimeKeepalivePop();
		callUserCallback(func);
	}, timeout);
};
var Browser = {
	mainLoop: {
		running: false,
		scheduler: null,
		method: '',
		currentlyRunningMainloop: 0,
		func: null,
		arg: 0,
		timingMode: 0,
		timingValue: 0,
		currentFrameNumber: 0,
		queue: [],
		pause() {
			Browser.mainLoop.scheduler = null;
			Browser.mainLoop.currentlyRunningMainloop++;
		},
		resume() {
			Browser.mainLoop.currentlyRunningMainloop++;
			var timingMode = Browser.mainLoop.timingMode;
			var timingValue = Browser.mainLoop.timingValue;
			var func = Browser.mainLoop.func;
			Browser.mainLoop.func = null;
			setMainLoop(func, 0, false, Browser.mainLoop.arg, true);
			_emscripten_set_main_loop_timing(timingMode, timingValue);
			Browser.mainLoop.scheduler();
		},
		updateStatus() {
			if (Module['setStatus']) {
				var message = Module['statusMessage'] || 'Please wait...';
				var remaining = Browser.mainLoop.remainingBlockers;
				var expected = Browser.mainLoop.expectedBlockers;
				if (remaining) {
					if (remaining < expected) {
						Module['setStatus'](message + ' (' + (expected - remaining) + '/' + expected + ')');
					} else {
						Module['setStatus'](message);
					}
				} else {
					Module['setStatus']('');
				}
			}
		},
		runIter(func) {
			if (ABORT) return;
			if (Module['preMainLoop']) {
				var preRet = Module['preMainLoop']();
				if (preRet === false) {
					return;
				}
			}
			callUserCallback(func);
			Module['postMainLoop']?.();
		}
	},
	isFullscreen: false,
	pointerLock: false,
	moduleContextCreatedCallbacks: [],
	workers: [],
	init() {
		if (Browser.initted) return;
		Browser.initted = true;
		var imagePlugin = {};
		imagePlugin['canHandle'] = function imagePlugin_canHandle(name) {
			return !Module.noImageDecoding && /\.(jpg|jpeg|png|bmp)$/i.test(name);
		};
		imagePlugin['handle'] = function imagePlugin_handle(byteArray, name, onload, onerror) {
			var b = new Blob([byteArray], { type: Browser.getMimetype(name) });
			if (b.size !== byteArray.length) {
				b = new Blob([new Uint8Array(byteArray).buffer], { type: Browser.getMimetype(name) });
			}
			var url = URL.createObjectURL(b);
			assert(typeof url == 'string', 'createObjectURL must return a url as a string');
			var img = new Image();
			img.onload = () => {
				assert(img.complete, `Image ${name} could not be decoded`);
				var canvas = document.createElement('canvas');
				canvas.width = img.width;
				canvas.height = img.height;
				var ctx = canvas.getContext('2d');
				ctx.drawImage(img, 0, 0);
				preloadedImages[name] = canvas;
				URL.revokeObjectURL(url);
				onload?.(byteArray);
			};
			img.onerror = (event) => {
				err(`Image ${url} could not be decoded`);
				onerror?.();
			};
			img.src = url;
		};
		preloadPlugins.push(imagePlugin);
		var audioPlugin = {};
		audioPlugin['canHandle'] = function audioPlugin_canHandle(name) {
			return !Module.noAudioDecoding && name.substr(-4) in { '.ogg': 1, '.wav': 1, '.mp3': 1 };
		};
		audioPlugin['handle'] = function audioPlugin_handle(byteArray, name, onload, onerror) {
			var done = false;
			function finish(audio) {
				if (done) return;
				done = true;
				preloadedAudios[name] = audio;
				onload?.(byteArray);
			}
			var b = new Blob([byteArray], { type: Browser.getMimetype(name) });
			var url = URL.createObjectURL(b);
			assert(typeof url == 'string', 'createObjectURL must return a url as a string');
			var audio = new Audio();
			audio.addEventListener('canplaythrough', () => finish(audio), false);
			audio.onerror = function audio_onerror(event) {
				if (done) return;
				err(`warning: browser could not fully decode audio ${name}, trying slower base64 approach`);
				function encode64(data) {
					var BASE = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';
					var PAD = '=';
					var ret = '';
					var leftchar = 0;
					var leftbits = 0;
					for (var i = 0; i < data.length; i++) {
						leftchar = (leftchar << 8) | data[i];
						leftbits += 8;
						while (leftbits >= 6) {
							var curr = (leftchar >> (leftbits - 6)) & 63;
							leftbits -= 6;
							ret += BASE[curr];
						}
					}
					if (leftbits == 2) {
						ret += BASE[(leftchar & 3) << 4];
						ret += PAD + PAD;
					} else if (leftbits == 4) {
						ret += BASE[(leftchar & 15) << 2];
						ret += PAD;
					}
					return ret;
				}
				audio.src = 'data:audio/x-' + name.substr(-3) + ';base64,' + encode64(byteArray);
				finish(audio);
			};
			audio.src = url;
			safeSetTimeout(() => {
				finish(audio);
			}, 1e4);
		};
		preloadPlugins.push(audioPlugin);
		function pointerLockChange() {
			Browser.pointerLock =
				document['pointerLockElement'] === Module['canvas'] ||
				document['mozPointerLockElement'] === Module['canvas'] ||
				document['webkitPointerLockElement'] === Module['canvas'] ||
				document['msPointerLockElement'] === Module['canvas'];
		}
		var canvas = Module['canvas'];
		if (canvas) {
			canvas.requestPointerLock =
				canvas['requestPointerLock'] ||
				canvas['mozRequestPointerLock'] ||
				canvas['webkitRequestPointerLock'] ||
				canvas['msRequestPointerLock'] ||
				(() => {});
			canvas.exitPointerLock =
				document['exitPointerLock'] ||
				document['mozExitPointerLock'] ||
				document['webkitExitPointerLock'] ||
				document['msExitPointerLock'] ||
				(() => {});
			canvas.exitPointerLock = canvas.exitPointerLock.bind(document);
			document.addEventListener('pointerlockchange', pointerLockChange, false);
			document.addEventListener('mozpointerlockchange', pointerLockChange, false);
			document.addEventListener('webkitpointerlockchange', pointerLockChange, false);
			document.addEventListener('mspointerlockchange', pointerLockChange, false);
			if (Module['elementPointerLock']) {
				canvas.addEventListener(
					'click',
					(ev) => {
						if (!Browser.pointerLock && Module['canvas'].requestPointerLock) {
							Module['canvas'].requestPointerLock();
							ev.preventDefault();
						}
					},
					false
				);
			}
		}
	},
	createContext(canvas, useWebGL, setInModule, webGLContextAttributes) {
		if (useWebGL && Module.ctx && canvas == Module.canvas) return Module.ctx;
		var ctx;
		var contextHandle;
		if (useWebGL) {
			var contextAttributes = { antialias: false, alpha: false, majorVersion: 1 };
			if (webGLContextAttributes) {
				for (var attribute in webGLContextAttributes) {
					contextAttributes[attribute] = webGLContextAttributes[attribute];
				}
			}
			if (typeof GL != 'undefined') {
				contextHandle = GL.createContext(canvas, contextAttributes);
				if (contextHandle) {
					ctx = GL.getContext(contextHandle).GLctx;
				}
			}
		} else {
			ctx = canvas.getContext('2d');
		}
		if (!ctx) return null;
		if (setInModule) {
			if (!useWebGL)
				assert(
					typeof GLctx == 'undefined',
					'cannot set in module if GLctx is used, but we are a non-GL context that would replace it'
				);
			Module.ctx = ctx;
			if (useWebGL) GL.makeContextCurrent(contextHandle);
			Module.useWebGL = useWebGL;
			Browser.moduleContextCreatedCallbacks.forEach((callback) => callback());
			Browser.init();
		}
		return ctx;
	},
	destroyContext(canvas, useWebGL, setInModule) {},
	fullscreenHandlersInstalled: false,
	lockPointer: undefined,
	resizeCanvas: undefined,
	requestFullscreen(lockPointer, resizeCanvas) {
		Browser.lockPointer = lockPointer;
		Browser.resizeCanvas = resizeCanvas;
		if (typeof Browser.lockPointer == 'undefined') Browser.lockPointer = true;
		if (typeof Browser.resizeCanvas == 'undefined') Browser.resizeCanvas = false;
		var canvas = Module['canvas'];
		function fullscreenChange() {
			Browser.isFullscreen = false;
			var canvasContainer = canvas.parentNode;
			if (
				(document['fullscreenElement'] ||
					document['mozFullScreenElement'] ||
					document['msFullscreenElement'] ||
					document['webkitFullscreenElement'] ||
					document['webkitCurrentFullScreenElement']) === canvasContainer
			) {
				canvas.exitFullscreen = Browser.exitFullscreen;
				if (Browser.lockPointer) canvas.requestPointerLock();
				Browser.isFullscreen = true;
				if (Browser.resizeCanvas) {
					Browser.setFullscreenCanvasSize();
				} else {
					Browser.updateCanvasDimensions(canvas);
				}
			} else {
				canvasContainer.parentNode.insertBefore(canvas, canvasContainer);
				canvasContainer.parentNode.removeChild(canvasContainer);
				if (Browser.resizeCanvas) {
					Browser.setWindowedCanvasSize();
				} else {
					Browser.updateCanvasDimensions(canvas);
				}
			}
			Module['onFullScreen']?.(Browser.isFullscreen);
			Module['onFullscreen']?.(Browser.isFullscreen);
		}
		if (!Browser.fullscreenHandlersInstalled) {
			Browser.fullscreenHandlersInstalled = true;
			document.addEventListener('fullscreenchange', fullscreenChange, false);
			document.addEventListener('mozfullscreenchange', fullscreenChange, false);
			document.addEventListener('webkitfullscreenchange', fullscreenChange, false);
			document.addEventListener('MSFullscreenChange', fullscreenChange, false);
		}
		var canvasContainer = document.createElement('div');
		canvas.parentNode.insertBefore(canvasContainer, canvas);
		canvasContainer.appendChild(canvas);
		canvasContainer.requestFullscreen =
			canvasContainer['requestFullscreen'] ||
			canvasContainer['mozRequestFullScreen'] ||
			canvasContainer['msRequestFullscreen'] ||
			(canvasContainer['webkitRequestFullscreen']
				? () => canvasContainer['webkitRequestFullscreen'](Element['ALLOW_KEYBOARD_INPUT'])
				: null) ||
			(canvasContainer['webkitRequestFullScreen']
				? () => canvasContainer['webkitRequestFullScreen'](Element['ALLOW_KEYBOARD_INPUT'])
				: null);
		canvasContainer.requestFullscreen();
	},
	requestFullScreen() {
		abort(
			'Module.requestFullScreen has been replaced by Module.requestFullscreen (without a capital S)'
		);
	},
	exitFullscreen() {
		if (!Browser.isFullscreen) {
			return false;
		}
		var CFS =
			document['exitFullscreen'] ||
			document['cancelFullScreen'] ||
			document['mozCancelFullScreen'] ||
			document['msExitFullscreen'] ||
			document['webkitCancelFullScreen'] ||
			(() => {});
		CFS.apply(document, []);
		return true;
	},
	nextRAF: 0,
	fakeRequestAnimationFrame(func) {
		var now = Date.now();
		if (Browser.nextRAF === 0) {
			Browser.nextRAF = now + 1e3 / 60;
		} else {
			while (now + 2 >= Browser.nextRAF) {
				Browser.nextRAF += 1e3 / 60;
			}
		}
		var delay = Math.max(Browser.nextRAF - now, 0);
		setTimeout(func, delay);
	},
	requestAnimationFrame(func) {
		if (typeof requestAnimationFrame == 'function') {
			requestAnimationFrame(func);
			return;
		}
		var RAF = Browser.fakeRequestAnimationFrame;
		RAF(func);
	},
	safeSetTimeout(func, timeout) {
		return safeSetTimeout(func, timeout);
	},
	safeRequestAnimationFrame(func) {
		runtimeKeepalivePush();
		return Browser.requestAnimationFrame(() => {
			runtimeKeepalivePop();
			callUserCallback(func);
		});
	},
	getMimetype(name) {
		return {
			jpg: 'image/jpeg',
			jpeg: 'image/jpeg',
			png: 'image/png',
			bmp: 'image/bmp',
			ogg: 'audio/ogg',
			wav: 'audio/wav',
			mp3: 'audio/mpeg'
		}[name.substr(name.lastIndexOf('.') + 1)];
	},
	getUserMedia(func) {
		window.getUserMedia ||= navigator['getUserMedia'] || navigator['mozGetUserMedia'];
		window.getUserMedia(func);
	},
	getMovementX(event) {
		return event['movementX'] || event['mozMovementX'] || event['webkitMovementX'] || 0;
	},
	getMovementY(event) {
		return event['movementY'] || event['mozMovementY'] || event['webkitMovementY'] || 0;
	},
	getMouseWheelDelta(event) {
		var delta = 0;
		switch (event.type) {
			case 'DOMMouseScroll':
				delta = event.detail / 3;
				break;
			case 'mousewheel':
				delta = event.wheelDelta / 120;
				break;
			case 'wheel':
				delta = event.deltaY;
				switch (event.deltaMode) {
					case 0:
						delta /= 100;
						break;
					case 1:
						delta /= 3;
						break;
					case 2:
						delta *= 80;
						break;
					default:
						throw 'unrecognized mouse wheel delta mode: ' + event.deltaMode;
				}
				break;
			default:
				throw 'unrecognized mouse wheel event: ' + event.type;
		}
		return delta;
	},
	mouseX: 0,
	mouseY: 0,
	mouseMovementX: 0,
	mouseMovementY: 0,
	touches: {},
	lastTouches: {},
	calculateMouseCoords(pageX, pageY) {
		var rect = Module['canvas'].getBoundingClientRect();
		var cw = Module['canvas'].width;
		var ch = Module['canvas'].height;
		var scrollX = typeof window.scrollX != 'undefined' ? window.scrollX : window.pageXOffset;
		var scrollY = typeof window.scrollY != 'undefined' ? window.scrollY : window.pageYOffset;
		assert(
			typeof scrollX != 'undefined' && typeof scrollY != 'undefined',
			'Unable to retrieve scroll position, mouse positions likely broken.'
		);
		var adjustedX = pageX - (scrollX + rect.left);
		var adjustedY = pageY - (scrollY + rect.top);
		adjustedX = adjustedX * (cw / rect.width);
		adjustedY = adjustedY * (ch / rect.height);
		return { x: adjustedX, y: adjustedY };
	},
	setMouseCoords(pageX, pageY) {
		const { x: x, y: y } = Browser.calculateMouseCoords(pageX, pageY);
		Browser.mouseMovementX = x - Browser.mouseX;
		Browser.mouseMovementY = y - Browser.mouseY;
		Browser.mouseX = x;
		Browser.mouseY = y;
	},
	calculateMouseEvent(event) {
		if (Browser.pointerLock) {
			if (event.type != 'mousemove' && 'mozMovementX' in event) {
				Browser.mouseMovementX = Browser.mouseMovementY = 0;
			} else {
				Browser.mouseMovementX = Browser.getMovementX(event);
				Browser.mouseMovementY = Browser.getMovementY(event);
			}
			if (typeof SDL != 'undefined') {
				Browser.mouseX = SDL.mouseX + Browser.mouseMovementX;
				Browser.mouseY = SDL.mouseY + Browser.mouseMovementY;
			} else {
				Browser.mouseX += Browser.mouseMovementX;
				Browser.mouseY += Browser.mouseMovementY;
			}
		} else {
			if (event.type === 'touchstart' || event.type === 'touchend' || event.type === 'touchmove') {
				var touch = event.touch;
				if (touch === undefined) {
					return;
				}
				var coords = Browser.calculateMouseCoords(touch.pageX, touch.pageY);
				if (event.type === 'touchstart') {
					Browser.lastTouches[touch.identifier] = coords;
					Browser.touches[touch.identifier] = coords;
				} else if (event.type === 'touchend' || event.type === 'touchmove') {
					var last = Browser.touches[touch.identifier];
					last ||= coords;
					Browser.lastTouches[touch.identifier] = last;
					Browser.touches[touch.identifier] = coords;
				}
				return;
			}
			Browser.setMouseCoords(event.pageX, event.pageY);
		}
	},
	resizeListeners: [],
	updateResizeListeners() {
		var canvas = Module['canvas'];
		Browser.resizeListeners.forEach((listener) => listener(canvas.width, canvas.height));
	},
	setCanvasSize(width, height, noUpdates) {
		var canvas = Module['canvas'];
		Browser.updateCanvasDimensions(canvas, width, height);
		if (!noUpdates) Browser.updateResizeListeners();
	},
	windowedWidth: 0,
	windowedHeight: 0,
	setFullscreenCanvasSize() {
		if (typeof SDL != 'undefined') {
			var flags = HEAPU32[SDL.screen >> 2];
			flags = flags | 8388608;
			HEAP32[SDL.screen >> 2] = flags;
		}
		Browser.updateCanvasDimensions(Module['canvas']);
		Browser.updateResizeListeners();
	},
	setWindowedCanvasSize() {
		if (typeof SDL != 'undefined') {
			var flags = HEAPU32[SDL.screen >> 2];
			flags = flags & ~8388608;
			HEAP32[SDL.screen >> 2] = flags;
		}
		Browser.updateCanvasDimensions(Module['canvas']);
		Browser.updateResizeListeners();
	},
	updateCanvasDimensions(canvas, wNative, hNative) {
		if (wNative && hNative) {
			canvas.widthNative = wNative;
			canvas.heightNative = hNative;
		} else {
			wNative = canvas.widthNative;
			hNative = canvas.heightNative;
		}
		var w = wNative;
		var h = hNative;
		if (Module['forcedAspectRatio'] && Module['forcedAspectRatio'] > 0) {
			if (w / h < Module['forcedAspectRatio']) {
				w = Math.round(h * Module['forcedAspectRatio']);
			} else {
				h = Math.round(w / Module['forcedAspectRatio']);
			}
		}
		if (
			(document['fullscreenElement'] ||
				document['mozFullScreenElement'] ||
				document['msFullscreenElement'] ||
				document['webkitFullscreenElement'] ||
				document['webkitCurrentFullScreenElement']) === canvas.parentNode &&
			typeof screen != 'undefined'
		) {
			var factor = Math.min(screen.width / w, screen.height / h);
			w = Math.round(w * factor);
			h = Math.round(h * factor);
		}
		if (Browser.resizeCanvas) {
			if (canvas.width != w) canvas.width = w;
			if (canvas.height != h) canvas.height = h;
			if (typeof canvas.style != 'undefined') {
				canvas.style.removeProperty('width');
				canvas.style.removeProperty('height');
			}
		} else {
			if (canvas.width != wNative) canvas.width = wNative;
			if (canvas.height != hNative) canvas.height = hNative;
			if (typeof canvas.style != 'undefined') {
				if (w != wNative || h != hNative) {
					canvas.style.setProperty('width', w + 'px', 'important');
					canvas.style.setProperty('height', h + 'px', 'important');
				} else {
					canvas.style.removeProperty('width');
					canvas.style.removeProperty('height');
				}
			}
		}
	}
};
var _emscripten_set_main_loop_timing = (mode, value) => {
	Browser.mainLoop.timingMode = mode;
	Browser.mainLoop.timingValue = value;
	if (!Browser.mainLoop.func) {
		err(
			'emscripten_set_main_loop_timing: Cannot set timing mode for main loop since a main loop does not exist! Call emscripten_set_main_loop first to set one up.'
		);
		return 1;
	}
	if (!Browser.mainLoop.running) {
		runtimeKeepalivePush();
		Browser.mainLoop.running = true;
	}
	if (mode == 0) {
		Browser.mainLoop.scheduler = function Browser_mainLoop_scheduler_setTimeout() {
			var timeUntilNextTick =
				Math.max(0, Browser.mainLoop.tickStartTime + value - _emscripten_get_now()) | 0;
			setTimeout(Browser.mainLoop.runner, timeUntilNextTick);
		};
		Browser.mainLoop.method = 'timeout';
	} else if (mode == 1) {
		Browser.mainLoop.scheduler = function Browser_mainLoop_scheduler_rAF() {
			Browser.requestAnimationFrame(Browser.mainLoop.runner);
		};
		Browser.mainLoop.method = 'rAF';
	} else if (mode == 2) {
		if (typeof Browser.setImmediate == 'undefined') {
			if (typeof setImmediate == 'undefined') {
				var setImmediates = [];
				var emscriptenMainLoopMessageId = 'setimmediate';
				var Browser_setImmediate_messageHandler = (event) => {
					if (
						event.data === emscriptenMainLoopMessageId ||
						event.data.target === emscriptenMainLoopMessageId
					) {
						event.stopPropagation();
						setImmediates.shift()();
					}
				};
				addEventListener('message', Browser_setImmediate_messageHandler, true);
				Browser.setImmediate = function Browser_emulated_setImmediate(func) {
					setImmediates.push(func);
					if (ENVIRONMENT_IS_WORKER) {
						if (Module['setImmediates'] === undefined) Module['setImmediates'] = [];
						Module['setImmediates'].push(func);
						postMessage({ target: emscriptenMainLoopMessageId });
					} else postMessage(emscriptenMainLoopMessageId, '*');
				};
			} else {
				Browser.setImmediate = setImmediate;
			}
		}
		Browser.mainLoop.scheduler = function Browser_mainLoop_scheduler_setImmediate() {
			Browser.setImmediate(Browser.mainLoop.runner);
		};
		Browser.mainLoop.method = 'immediate';
	}
	return 0;
};
var setMainLoop = (browserIterationFunc, fps, simulateInfiniteLoop, arg, noSetTiming) => {
	assert(
		!Browser.mainLoop.func,
		'emscripten_set_main_loop: there can only be one main loop function at once: call emscripten_cancel_main_loop to cancel the previous one before setting a new one with different parameters.'
	);
	Browser.mainLoop.func = browserIterationFunc;
	Browser.mainLoop.arg = arg;
	var thisMainLoopId = Browser.mainLoop.currentlyRunningMainloop;
	function checkIsRunning() {
		if (thisMainLoopId < Browser.mainLoop.currentlyRunningMainloop) {
			runtimeKeepalivePop();
			return false;
		}
		return true;
	}
	Browser.mainLoop.running = false;
	Browser.mainLoop.runner = function Browser_mainLoop_runner() {
		if (ABORT) return;
		if (Browser.mainLoop.queue.length > 0) {
			var start = Date.now();
			var blocker = Browser.mainLoop.queue.shift();
			blocker.func(blocker.arg);
			if (Browser.mainLoop.remainingBlockers) {
				var remaining = Browser.mainLoop.remainingBlockers;
				var next = remaining % 1 == 0 ? remaining - 1 : Math.floor(remaining);
				if (blocker.counted) {
					Browser.mainLoop.remainingBlockers = next;
				} else {
					next = next + 0.5;
					Browser.mainLoop.remainingBlockers = (8 * remaining + next) / 9;
				}
			}
			Browser.mainLoop.updateStatus();
			if (!checkIsRunning()) return;
			setTimeout(Browser.mainLoop.runner, 0);
			return;
		}
		if (!checkIsRunning()) return;
		Browser.mainLoop.currentFrameNumber = (Browser.mainLoop.currentFrameNumber + 1) | 0;
		if (
			Browser.mainLoop.timingMode == 1 &&
			Browser.mainLoop.timingValue > 1 &&
			Browser.mainLoop.currentFrameNumber % Browser.mainLoop.timingValue != 0
		) {
			Browser.mainLoop.scheduler();
			return;
		} else if (Browser.mainLoop.timingMode == 0) {
			Browser.mainLoop.tickStartTime = _emscripten_get_now();
		}
		if (Browser.mainLoop.method === 'timeout' && Module.ctx) {
			warnOnce(
				'Looks like you are rendering without using requestAnimationFrame for the main loop. You should use 0 for the frame rate in emscripten_set_main_loop in order to use requestAnimationFrame, as that can greatly improve your frame rates!'
			);
			Browser.mainLoop.method = '';
		}
		Browser.mainLoop.runIter(browserIterationFunc);
		checkStackCookie();
		if (!checkIsRunning()) return;
		if (typeof SDL == 'object') SDL.audio?.queueNewAudioData?.();
		Browser.mainLoop.scheduler();
	};
	if (!noSetTiming) {
		if (fps && fps > 0) {
			_emscripten_set_main_loop_timing(0, 1e3 / fps);
		} else {
			_emscripten_set_main_loop_timing(1, 1);
		}
		Browser.mainLoop.scheduler();
	}
	if (simulateInfiniteLoop) {
		throw 'unwind';
	}
};
var _emscripten_set_main_loop = (func, fps, simulateInfiniteLoop) => {
	var browserIterationFunc = () => dynCall_v.call(null, func);
	setMainLoop(browserIterationFunc, fps, simulateInfiniteLoop);
};
var fillPointerlockChangeEventData = (eventStruct) => {
	var pointerLockElement =
		document.pointerLockElement ||
		document.mozPointerLockElement ||
		document.webkitPointerLockElement ||
		document.msPointerLockElement;
	var isPointerlocked = !!pointerLockElement;
	HEAP32[eventStruct >> 2] = isPointerlocked;
	var nodeName = JSEvents.getNodeNameForTarget(pointerLockElement);
	var id = pointerLockElement?.id || '';
	stringToUTF8(nodeName, eventStruct + 4, 128);
	stringToUTF8(id, eventStruct + 132, 128);
};
var registerPointerlockChangeEventCallback = (
	target,
	userData,
	useCapture,
	callbackfunc,
	eventTypeId,
	eventTypeString,
	targetThread
) => {
	targetThread = JSEvents.getTargetThreadForEventCallback(targetThread);
	if (!JSEvents.pointerlockChangeEvent) JSEvents.pointerlockChangeEvent = _malloc(260);
	var pointerlockChangeEventHandlerFunc = (e = event) => {
		var pointerlockChangeEvent = targetThread ? _malloc(260) : JSEvents.pointerlockChangeEvent;
		fillPointerlockChangeEventData(pointerlockChangeEvent);
		if (targetThread)
			__emscripten_run_callback_on_thread(
				targetThread,
				callbackfunc,
				eventTypeId,
				pointerlockChangeEvent,
				userData
			);
		else if (
			((a1, a2, a3) => dynCall_iiii.apply(null, [callbackfunc, a1, a2, a3]))(
				eventTypeId,
				pointerlockChangeEvent,
				userData
			)
		)
			e.preventDefault();
	};
	var eventHandler = {
		target: target,
		eventTypeString: eventTypeString,
		callbackfunc: callbackfunc,
		handlerFunc: pointerlockChangeEventHandlerFunc,
		useCapture: useCapture
	};
	return JSEvents.registerOrRemoveHandler(eventHandler);
};
function _emscripten_set_pointerlockchange_callback_on_thread(
	target,
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(17, 1, target, userData, useCapture, callbackfunc, targetThread);
	if (
		!document ||
		!document.body ||
		(!document.body.requestPointerLock &&
			!document.body.mozRequestPointerLock &&
			!document.body.webkitRequestPointerLock &&
			!document.body.msRequestPointerLock)
	) {
		return -1;
	}
	target = findEventTarget(target);
	if (!target) return -4;
	registerPointerlockChangeEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		20,
		'mozpointerlockchange',
		targetThread
	);
	registerPointerlockChangeEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		20,
		'webkitpointerlockchange',
		targetThread
	);
	registerPointerlockChangeEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		20,
		'mspointerlockchange',
		targetThread
	);
	return registerPointerlockChangeEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		20,
		'pointerlockchange',
		targetThread
	);
}
var registerUiEventCallback = (
	target,
	userData,
	useCapture,
	callbackfunc,
	eventTypeId,
	eventTypeString,
	targetThread
) => {
	targetThread = JSEvents.getTargetThreadForEventCallback(targetThread);
	if (!JSEvents.uiEvent) JSEvents.uiEvent = _malloc(36);
	target = findEventTarget(target);
	var uiEventHandlerFunc = (e = event) => {
		if (e.target != target) {
			return;
		}
		var b = document.body;
		if (!b) {
			return;
		}
		var uiEvent = targetThread ? _malloc(36) : JSEvents.uiEvent;
		HEAP32[uiEvent >> 2] = e.detail;
		HEAP32[(uiEvent + 4) >> 2] = b.clientWidth;
		HEAP32[(uiEvent + 8) >> 2] = b.clientHeight;
		HEAP32[(uiEvent + 12) >> 2] = innerWidth;
		HEAP32[(uiEvent + 16) >> 2] = innerHeight;
		HEAP32[(uiEvent + 20) >> 2] = outerWidth;
		HEAP32[(uiEvent + 24) >> 2] = outerHeight;
		HEAP32[(uiEvent + 28) >> 2] = pageXOffset;
		HEAP32[(uiEvent + 32) >> 2] = pageYOffset;
		if (targetThread)
			__emscripten_run_callback_on_thread(
				targetThread,
				callbackfunc,
				eventTypeId,
				uiEvent,
				userData
			);
		else if (
			((a1, a2, a3) => dynCall_iiii.apply(null, [callbackfunc, a1, a2, a3]))(
				eventTypeId,
				uiEvent,
				userData
			)
		)
			e.preventDefault();
	};
	var eventHandler = {
		target: target,
		eventTypeString: eventTypeString,
		callbackfunc: callbackfunc,
		handlerFunc: uiEventHandlerFunc,
		useCapture: useCapture
	};
	return JSEvents.registerOrRemoveHandler(eventHandler);
};
function _emscripten_set_resize_callback_on_thread(
	target,
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(18, 1, target, userData, useCapture, callbackfunc, targetThread);
	return registerUiEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		10,
		'resize',
		targetThread
	);
}
var registerTouchEventCallback = (
	target,
	userData,
	useCapture,
	callbackfunc,
	eventTypeId,
	eventTypeString,
	targetThread
) => {
	targetThread = JSEvents.getTargetThreadForEventCallback(targetThread);
	if (!JSEvents.touchEvent) JSEvents.touchEvent = _malloc(1696);
	target = findEventTarget(target);
	var touchEventHandlerFunc = (e) => {
		assert(e);
		var t,
			touches = {},
			et = e.touches;
		for (var i = 0; i < et.length; ++i) {
			t = et[i];
			t.isChanged = t.onTarget = 0;
			touches[t.identifier] = t;
		}
		for (var i = 0; i < e.changedTouches.length; ++i) {
			t = e.changedTouches[i];
			t.isChanged = 1;
			touches[t.identifier] = t;
		}
		for (var i = 0; i < e.targetTouches.length; ++i) {
			touches[e.targetTouches[i].identifier].onTarget = 1;
		}
		var touchEvent = targetThread ? _malloc(1696) : JSEvents.touchEvent;
		HEAPF64[touchEvent >> 3] = e.timeStamp;
		var idx = touchEvent >> 2;
		HEAP32[idx + 3] = e.ctrlKey;
		HEAP32[idx + 4] = e.shiftKey;
		HEAP32[idx + 5] = e.altKey;
		HEAP32[idx + 6] = e.metaKey;
		idx += 7;
		var targetRect = getBoundingClientRect(target);
		var numTouches = 0;
		for (var i in touches) {
			t = touches[i];
			HEAP32[idx + 0] = t.identifier;
			HEAP32[idx + 1] = t.screenX;
			HEAP32[idx + 2] = t.screenY;
			HEAP32[idx + 3] = t.clientX;
			HEAP32[idx + 4] = t.clientY;
			HEAP32[idx + 5] = t.pageX;
			HEAP32[idx + 6] = t.pageY;
			HEAP32[idx + 7] = t.isChanged;
			HEAP32[idx + 8] = t.onTarget;
			HEAP32[idx + 9] = t.clientX - targetRect.left;
			HEAP32[idx + 10] = t.clientY - targetRect.top;
			idx += 13;
			if (++numTouches > 31) {
				break;
			}
		}
		HEAP32[(touchEvent + 8) >> 2] = numTouches;
		if (targetThread)
			__emscripten_run_callback_on_thread(
				targetThread,
				callbackfunc,
				eventTypeId,
				touchEvent,
				userData
			);
		else if (
			((a1, a2, a3) => dynCall_iiii.apply(null, [callbackfunc, a1, a2, a3]))(
				eventTypeId,
				touchEvent,
				userData
			)
		)
			e.preventDefault();
	};
	var eventHandler = {
		target: target,
		allowsDeferredCalls: eventTypeString == 'touchstart' || eventTypeString == 'touchend',
		eventTypeString: eventTypeString,
		callbackfunc: callbackfunc,
		handlerFunc: touchEventHandlerFunc,
		useCapture: useCapture
	};
	return JSEvents.registerOrRemoveHandler(eventHandler);
};
function _emscripten_set_touchcancel_callback_on_thread(
	target,
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(19, 1, target, userData, useCapture, callbackfunc, targetThread);
	return registerTouchEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		25,
		'touchcancel',
		targetThread
	);
}
function _emscripten_set_touchend_callback_on_thread(
	target,
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(20, 1, target, userData, useCapture, callbackfunc, targetThread);
	return registerTouchEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		23,
		'touchend',
		targetThread
	);
}
function _emscripten_set_touchmove_callback_on_thread(
	target,
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(21, 1, target, userData, useCapture, callbackfunc, targetThread);
	return registerTouchEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		24,
		'touchmove',
		targetThread
	);
}
function _emscripten_set_touchstart_callback_on_thread(
	target,
	userData,
	useCapture,
	callbackfunc,
	targetThread
) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(22, 1, target, userData, useCapture, callbackfunc, targetThread);
	return registerTouchEventCallback(
		target,
		userData,
		useCapture,
		callbackfunc,
		22,
		'touchstart',
		targetThread
	);
}
function _emscripten_set_window_title(title) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(23, 1, title);
	return (document.title = UTF8ToString(title));
}
function _fd_close(fd) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(24, 1, fd);
	try {
		var stream = SYSCALLS.getStreamFromFD(fd);
		FS.close(stream);
		return 0;
	} catch (e) {
		if (typeof FS == 'undefined' || !(e.name === 'ErrnoError')) throw e;
		return e.errno;
	}
}
var doReadv = (stream, iov, iovcnt, offset) => {
	var ret = 0;
	for (var i = 0; i < iovcnt; i++) {
		var ptr = HEAPU32[iov >> 2];
		var len = HEAPU32[(iov + 4) >> 2];
		iov += 8;
		var curr = FS.read(stream, HEAP8, ptr, len, offset);
		if (curr < 0) return -1;
		ret += curr;
		if (curr < len) break;
		if (typeof offset !== 'undefined') {
			offset += curr;
		}
	}
	return ret;
};
function _fd_read(fd, iov, iovcnt, pnum) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(25, 1, fd, iov, iovcnt, pnum);
	try {
		var stream = SYSCALLS.getStreamFromFD(fd);
		var num = doReadv(stream, iov, iovcnt);
		HEAPU32[pnum >> 2] = num;
		return 0;
	} catch (e) {
		if (typeof FS == 'undefined' || !(e.name === 'ErrnoError')) throw e;
		return e.errno;
	}
}
function _fd_seek(fd, offset_low, offset_high, whence, newOffset) {
	if (ENVIRONMENT_IS_PTHREAD)
		return proxyToMainThread(26, 1, fd, offset_low, offset_high, whence, newOffset);
	var offset = convertI32PairToI53Checked(offset_low, offset_high);
	try {
		if (isNaN(offset)) return 61;
		var stream = SYSCALLS.getStreamFromFD(fd);
		FS.llseek(stream, offset, whence);
		(tempI64 = [
			stream.position >>> 0,
			((tempDouble = stream.position),
			+Math.abs(tempDouble) >= 1
				? tempDouble > 0
					? +Math.floor(tempDouble / 4294967296) >>> 0
					: ~~+Math.ceil((tempDouble - +(~~tempDouble >>> 0)) / 4294967296) >>> 0
				: 0)
		]),
			(HEAP32[newOffset >> 2] = tempI64[0]),
			(HEAP32[(newOffset + 4) >> 2] = tempI64[1]);
		if (stream.getdents && offset === 0 && whence === 0) stream.getdents = null;
		return 0;
	} catch (e) {
		if (typeof FS == 'undefined' || !(e.name === 'ErrnoError')) throw e;
		return e.errno;
	}
}
var doWritev = (stream, iov, iovcnt, offset) => {
	var ret = 0;
	for (var i = 0; i < iovcnt; i++) {
		var ptr = HEAPU32[iov >> 2];
		var len = HEAPU32[(iov + 4) >> 2];
		iov += 8;
		var curr = FS.write(stream, HEAP8, ptr, len, offset);
		if (curr < 0) return -1;
		ret += curr;
		if (typeof offset !== 'undefined') {
			offset += curr;
		}
	}
	return ret;
};
function _fd_write(fd, iov, iovcnt, pnum) {
	if (ENVIRONMENT_IS_PTHREAD) return proxyToMainThread(27, 1, fd, iov, iovcnt, pnum);
	try {
		var stream = SYSCALLS.getStreamFromFD(fd);
		var num = doWritev(stream, iov, iovcnt);
		HEAPU32[pnum >> 2] = num;
		return 0;
	} catch (e) {
		if (typeof FS == 'undefined' || !(e.name === 'ErrnoError')) throw e;
		return e.errno;
	}
}
var _getentropy = (buffer, size) => {
	randomFill(HEAPU8.subarray(buffer, buffer + size));
	return 0;
};
function GLFW_Window(
	id,
	width,
	height,
	framebufferWidth,
	framebufferHeight,
	title,
	monitor,
	share
) {
	this.id = id;
	this.x = 0;
	this.y = 0;
	this.fullscreen = false;
	this.storedX = 0;
	this.storedY = 0;
	this.width = width;
	this.height = height;
	this.framebufferWidth = framebufferWidth;
	this.framebufferHeight = framebufferHeight;
	this.storedWidth = width;
	this.storedHeight = height;
	this.title = title;
	this.monitor = monitor;
	this.share = share;
	this.attributes = Object.assign({}, GLFW.hints);
	this.inputModes = { 208897: 212993, 208898: 0, 208899: 0 };
	this.buttons = 0;
	this.keys = new Array();
	this.domKeys = new Array();
	this.shouldClose = 0;
	this.title = null;
	this.windowPosFunc = 0;
	this.windowSizeFunc = 0;
	this.windowCloseFunc = 0;
	this.windowRefreshFunc = 0;
	this.windowFocusFunc = 0;
	this.windowIconifyFunc = 0;
	this.windowMaximizeFunc = 0;
	this.framebufferSizeFunc = 0;
	this.windowContentScaleFunc = 0;
	this.mouseButtonFunc = 0;
	this.cursorPosFunc = 0;
	this.cursorEnterFunc = 0;
	this.scrollFunc = 0;
	this.dropFunc = 0;
	this.keyFunc = 0;
	this.charFunc = 0;
	this.userptr = 0;
}
var GLFW = {
	WindowFromId: (id) => {
		if (id <= 0 || !GLFW.windows) return null;
		return GLFW.windows[id - 1];
	},
	joystickFunc: 0,
	errorFunc: 0,
	monitorFunc: 0,
	active: null,
	scale: null,
	windows: null,
	monitors: null,
	monitorString: null,
	versionString: null,
	initialTime: null,
	extensions: null,
	devicePixelRatioMQL: null,
	hints: null,
	primaryTouchId: null,
	defaultHints: {
		131073: 0,
		131074: 0,
		131075: 1,
		131076: 1,
		131077: 1,
		131082: 0,
		135169: 8,
		135170: 8,
		135171: 8,
		135172: 8,
		135173: 24,
		135174: 8,
		135175: 0,
		135176: 0,
		135177: 0,
		135178: 0,
		135179: 0,
		135180: 0,
		135181: 0,
		135182: 0,
		135183: 0,
		139265: 196609,
		139266: 1,
		139267: 0,
		139268: 0,
		139269: 0,
		139270: 0,
		139271: 0,
		139272: 0,
		139276: 0
	},
	DOMToGLFWKeyCode: (keycode) => {
		switch (keycode) {
			case 32:
				return 32;
			case 222:
				return 39;
			case 188:
				return 44;
			case 173:
				return 45;
			case 189:
				return 45;
			case 190:
				return 46;
			case 191:
				return 47;
			case 48:
				return 48;
			case 49:
				return 49;
			case 50:
				return 50;
			case 51:
				return 51;
			case 52:
				return 52;
			case 53:
				return 53;
			case 54:
				return 54;
			case 55:
				return 55;
			case 56:
				return 56;
			case 57:
				return 57;
			case 59:
				return 59;
			case 61:
				return 61;
			case 187:
				return 61;
			case 65:
				return 65;
			case 66:
				return 66;
			case 67:
				return 67;
			case 68:
				return 68;
			case 69:
				return 69;
			case 70:
				return 70;
			case 71:
				return 71;
			case 72:
				return 72;
			case 73:
				return 73;
			case 74:
				return 74;
			case 75:
				return 75;
			case 76:
				return 76;
			case 77:
				return 77;
			case 78:
				return 78;
			case 79:
				return 79;
			case 80:
				return 80;
			case 81:
				return 81;
			case 82:
				return 82;
			case 83:
				return 83;
			case 84:
				return 84;
			case 85:
				return 85;
			case 86:
				return 86;
			case 87:
				return 87;
			case 88:
				return 88;
			case 89:
				return 89;
			case 90:
				return 90;
			case 219:
				return 91;
			case 220:
				return 92;
			case 221:
				return 93;
			case 192:
				return 96;
			case 27:
				return 256;
			case 13:
				return 257;
			case 9:
				return 258;
			case 8:
				return 259;
			case 45:
				return 260;
			case 46:
				return 261;
			case 39:
				return 262;
			case 37:
				return 263;
			case 40:
				return 264;
			case 38:
				return 265;
			case 33:
				return 266;
			case 34:
				return 267;
			case 36:
				return 268;
			case 35:
				return 269;
			case 20:
				return 280;
			case 145:
				return 281;
			case 144:
				return 282;
			case 44:
				return 283;
			case 19:
				return 284;
			case 112:
				return 290;
			case 113:
				return 291;
			case 114:
				return 292;
			case 115:
				return 293;
			case 116:
				return 294;
			case 117:
				return 295;
			case 118:
				return 296;
			case 119:
				return 297;
			case 120:
				return 298;
			case 121:
				return 299;
			case 122:
				return 300;
			case 123:
				return 301;
			case 124:
				return 302;
			case 125:
				return 303;
			case 126:
				return 304;
			case 127:
				return 305;
			case 128:
				return 306;
			case 129:
				return 307;
			case 130:
				return 308;
			case 131:
				return 309;
			case 132:
				return 310;
			case 133:
				return 311;
			case 134:
				return 312;
			case 135:
				return 313;
			case 136:
				return 314;
			case 96:
				return 320;
			case 97:
				return 321;
			case 98:
				return 322;
			case 99:
				return 323;
			case 100:
				return 324;
			case 101:
				return 325;
			case 102:
				return 326;
			case 103:
				return 327;
			case 104:
				return 328;
			case 105:
				return 329;
			case 110:
				return 330;
			case 111:
				return 331;
			case 106:
				return 332;
			case 109:
				return 333;
			case 107:
				return 334;
			case 16:
				return 340;
			case 17:
				return 341;
			case 18:
				return 342;
			case 91:
				return 343;
			case 93:
				return 348;
			default:
				return -1;
		}
	},
	getModBits: (win) => {
		var mod = 0;
		if (win.keys[340]) mod |= 1;
		if (win.keys[341]) mod |= 2;
		if (win.keys[342]) mod |= 4;
		if (win.keys[343]) mod |= 8;
		return mod;
	},
	onKeyPress: (event) => {
		if (!GLFW.active || !GLFW.active.charFunc) return;
		if (event.ctrlKey || event.metaKey) return;
		var charCode = event.charCode;
		if (charCode == 0 || (charCode >= 0 && charCode <= 31)) return;
		((a1, a2) => dynCall_vii.apply(null, [GLFW.active.charFunc, a1, a2]))(GLFW.active.id, charCode);
	},
	onKeyChanged: (keyCode, status) => {
		if (!GLFW.active) return;
		var key = GLFW.DOMToGLFWKeyCode(keyCode);
		if (key == -1) return;
		var repeat = status && GLFW.active.keys[key];
		GLFW.active.keys[key] = status;
		GLFW.active.domKeys[keyCode] = status;
		if (GLFW.active.keyFunc) {
			if (repeat) status = 2;
			((a1, a2, a3, a4, a5) =>
				dynCall_viiiii.apply(null, [GLFW.active.keyFunc, a1, a2, a3, a4, a5]))(
				GLFW.active.id,
				key,
				keyCode,
				status,
				GLFW.getModBits(GLFW.active)
			);
		}
	},
	onGamepadConnected: (event) => {
		GLFW.refreshJoysticks();
	},
	onGamepadDisconnected: (event) => {
		GLFW.refreshJoysticks();
	},
	onKeydown: (event) => {
		GLFW.onKeyChanged(event.keyCode, 1);
		if (event.keyCode === 8 || event.keyCode === 9) {
			event.preventDefault();
		}
	},
	onKeyup: (event) => {
		GLFW.onKeyChanged(event.keyCode, 0);
	},
	onBlur: (event) => {
		if (!GLFW.active) return;
		for (var i = 0; i < GLFW.active.domKeys.length; ++i) {
			if (GLFW.active.domKeys[i]) {
				GLFW.onKeyChanged(i, 0);
			}
		}
	},
	onMousemove: (event) => {
		if (!GLFW.active) return;
		if (event.type === 'touchmove') {
			event.preventDefault();
			let primaryChanged = false;
			for (let i of event.changedTouches) {
				if (GLFW.primaryTouchId === i.identifier) {
					Browser.setMouseCoords(i.pageX, i.pageY);
					primaryChanged = true;
					break;
				}
			}
			if (!primaryChanged) {
				return;
			}
		} else {
			Browser.calculateMouseEvent(event);
		}
		if (event.target != Module['canvas'] || !GLFW.active.cursorPosFunc) return;
		if (GLFW.active.cursorPosFunc) {
			((a1, a2, a3) => dynCall_vidd.apply(null, [GLFW.active.cursorPosFunc, a1, a2, a3]))(
				GLFW.active.id,
				Browser.mouseX,
				Browser.mouseY
			);
		}
	},
	DOMToGLFWMouseButton: (event) => {
		var eventButton = event['button'];
		if (eventButton > 0) {
			if (eventButton == 1) {
				eventButton = 2;
			} else {
				eventButton = 1;
			}
		}
		return eventButton;
	},
	onMouseenter: (event) => {
		if (!GLFW.active) return;
		if (event.target != Module['canvas']) return;
		if (GLFW.active.cursorEnterFunc) {
			((a1, a2) => dynCall_vii.apply(null, [GLFW.active.cursorEnterFunc, a1, a2]))(
				GLFW.active.id,
				1
			);
		}
	},
	onMouseleave: (event) => {
		if (!GLFW.active) return;
		if (event.target != Module['canvas']) return;
		if (GLFW.active.cursorEnterFunc) {
			((a1, a2) => dynCall_vii.apply(null, [GLFW.active.cursorEnterFunc, a1, a2]))(
				GLFW.active.id,
				0
			);
		}
	},
	onMouseButtonChanged: (event, status) => {
		if (!GLFW.active) return;
		if (event.target != Module['canvas']) return;
		const isTouchType =
			event.type === 'touchstart' || event.type === 'touchend' || event.type === 'touchcancel';
		let eventButton = 0;
		if (isTouchType) {
			event.preventDefault();
			let primaryChanged = false;
			if (
				GLFW.primaryTouchId === null &&
				event.type === 'touchstart' &&
				event.targetTouches.length > 0
			) {
				const chosenTouch = event.targetTouches[0];
				GLFW.primaryTouchId = chosenTouch.identifier;
				Browser.setMouseCoords(chosenTouch.pageX, chosenTouch.pageY);
				primaryChanged = true;
			} else if (event.type === 'touchend' || event.type === 'touchcancel') {
				for (let i of event.changedTouches) {
					if (GLFW.primaryTouchId === i.identifier) {
						GLFW.primaryTouchId = null;
						primaryChanged = true;
						break;
					}
				}
			}
			if (!primaryChanged) {
				return;
			}
		} else {
			Browser.calculateMouseEvent(event);
			eventButton = GLFW.DOMToGLFWMouseButton(event);
		}
		if (status == 1) {
			GLFW.active.buttons |= 1 << eventButton;
			try {
				event.target.setCapture();
			} catch (e) {}
		} else {
			GLFW.active.buttons &= ~(1 << eventButton);
		}
		if (GLFW.active.mouseButtonFunc) {
			((a1, a2, a3, a4) =>
				dynCall_viiii.apply(null, [GLFW.active.mouseButtonFunc, a1, a2, a3, a4]))(
				GLFW.active.id,
				eventButton,
				status,
				GLFW.getModBits(GLFW.active)
			);
		}
	},
	onMouseButtonDown: (event) => {
		if (!GLFW.active) return;
		GLFW.onMouseButtonChanged(event, 1);
	},
	onMouseButtonUp: (event) => {
		if (!GLFW.active) return;
		GLFW.onMouseButtonChanged(event, 0);
	},
	onMouseWheel: (event) => {
		var delta = -Browser.getMouseWheelDelta(event);
		delta = delta == 0 ? 0 : delta > 0 ? Math.max(delta, 1) : Math.min(delta, -1);
		GLFW.wheelPos += delta;
		if (!GLFW.active || !GLFW.active.scrollFunc || event.target != Module['canvas']) return;
		var sx = 0;
		var sy = delta;
		if (event.type == 'mousewheel') {
			sx = event.wheelDeltaX;
		} else {
			sx = event.deltaX;
		}
		((a1, a2, a3) => dynCall_vidd.apply(null, [GLFW.active.scrollFunc, a1, a2, a3]))(
			GLFW.active.id,
			sx,
			sy
		);
		event.preventDefault();
	},
	onCanvasResize: (width, height, framebufferWidth, framebufferHeight) => {
		if (!GLFW.active) return;
		var resizeNeeded = false;
		if (
			document['fullscreen'] ||
			document['fullScreen'] ||
			document['mozFullScreen'] ||
			document['webkitIsFullScreen']
		) {
			if (!GLFW.active.fullscreen) {
				resizeNeeded = width != screen.width || height != screen.height;
				GLFW.active.storedX = GLFW.active.x;
				GLFW.active.storedY = GLFW.active.y;
				GLFW.active.storedWidth = GLFW.active.width;
				GLFW.active.storedHeight = GLFW.active.height;
				GLFW.active.x = GLFW.active.y = 0;
				GLFW.active.width = screen.width;
				GLFW.active.height = screen.height;
				GLFW.active.fullscreen = true;
			}
		} else if (GLFW.active.fullscreen == true) {
			resizeNeeded = width != GLFW.active.storedWidth || height != GLFW.active.storedHeight;
			GLFW.active.x = GLFW.active.storedX;
			GLFW.active.y = GLFW.active.storedY;
			GLFW.active.width = GLFW.active.storedWidth;
			GLFW.active.height = GLFW.active.storedHeight;
			GLFW.active.fullscreen = false;
		}
		if (resizeNeeded) {
			Browser.setCanvasSize(GLFW.active.width, GLFW.active.height);
		} else if (
			GLFW.active.width != width ||
			GLFW.active.height != height ||
			GLFW.active.framebufferWidth != framebufferWidth ||
			GLFW.active.framebufferHeight != framebufferHeight
		) {
			GLFW.active.width = width;
			GLFW.active.height = height;
			GLFW.active.framebufferWidth = framebufferWidth;
			GLFW.active.framebufferHeight = framebufferHeight;
			GLFW.onWindowSizeChanged();
			GLFW.onFramebufferSizeChanged();
		}
	},
	onWindowSizeChanged: () => {
		if (!GLFW.active) return;
		if (GLFW.active.windowSizeFunc) {
			((a1, a2, a3) => dynCall_viii.apply(null, [GLFW.active.windowSizeFunc, a1, a2, a3]))(
				GLFW.active.id,
				GLFW.active.width,
				GLFW.active.height
			);
		}
	},
	onFramebufferSizeChanged: () => {
		if (!GLFW.active) return;
		if (GLFW.active.framebufferSizeFunc) {
			((a1, a2, a3) => dynCall_viii.apply(null, [GLFW.active.framebufferSizeFunc, a1, a2, a3]))(
				GLFW.active.id,
				GLFW.active.framebufferWidth,
				GLFW.active.framebufferHeight
			);
		}
	},
	onWindowContentScaleChanged: (scale) => {
		GLFW.scale = scale;
		if (!GLFW.active) return;
		if (GLFW.active.windowContentScaleFunc) {
			((a1, a2, a3) => dynCall_viff.apply(null, [GLFW.active.windowContentScaleFunc, a1, a2, a3]))(
				GLFW.active.id,
				GLFW.scale,
				GLFW.scale
			);
		}
	},
	getTime: () => _emscripten_get_now() / 1e3,
	setWindowTitle: (winid, title) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return;
		win.title = title;
		if (GLFW.active.id == win.id) {
			_emscripten_set_window_title(title);
		}
	},
	setJoystickCallback: (cbfun) => {
		var prevcbfun = GLFW.joystickFunc;
		GLFW.joystickFunc = cbfun;
		GLFW.refreshJoysticks();
		return prevcbfun;
	},
	joys: {},
	lastGamepadState: [],
	lastGamepadStateFrame: null,
	refreshJoysticks: () => {
		if (
			Browser.mainLoop.currentFrameNumber !== GLFW.lastGamepadStateFrame ||
			!Browser.mainLoop.currentFrameNumber
		) {
			GLFW.lastGamepadState = navigator.getGamepads
				? navigator.getGamepads()
				: navigator.webkitGetGamepads || [];
			GLFW.lastGamepadStateFrame = Browser.mainLoop.currentFrameNumber;
			for (var joy = 0; joy < GLFW.lastGamepadState.length; ++joy) {
				var gamepad = GLFW.lastGamepadState[joy];
				if (gamepad) {
					if (!GLFW.joys[joy]) {
						out('glfw joystick connected:', joy);
						GLFW.joys[joy] = {
							id: stringToNewUTF8(gamepad.id),
							buttonsCount: gamepad.buttons.length,
							axesCount: gamepad.axes.length,
							buttons: _malloc(gamepad.buttons.length),
							axes: _malloc(gamepad.axes.length * 4)
						};
						if (GLFW.joystickFunc) {
							((a1, a2) => dynCall_vii.apply(null, [GLFW.joystickFunc, a1, a2]))(joy, 262145);
						}
					}
					var data = GLFW.joys[joy];
					for (var i = 0; i < gamepad.buttons.length; ++i) {
						HEAP8[(data.buttons + i) >> 0] = gamepad.buttons[i].pressed;
					}
					for (var i = 0; i < gamepad.axes.length; ++i) {
						HEAPF32[(data.axes + i * 4) >> 2] = gamepad.axes[i];
					}
				} else {
					if (GLFW.joys[joy]) {
						out('glfw joystick disconnected', joy);
						if (GLFW.joystickFunc) {
							((a1, a2) => dynCall_vii.apply(null, [GLFW.joystickFunc, a1, a2]))(joy, 262146);
						}
						_free(GLFW.joys[joy].id);
						_free(GLFW.joys[joy].buttons);
						_free(GLFW.joys[joy].axes);
						delete GLFW.joys[joy];
					}
				}
			}
		}
	},
	setKeyCallback: (winid, cbfun) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return null;
		var prevcbfun = win.keyFunc;
		win.keyFunc = cbfun;
		return prevcbfun;
	},
	setCharCallback: (winid, cbfun) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return null;
		var prevcbfun = win.charFunc;
		win.charFunc = cbfun;
		return prevcbfun;
	},
	setMouseButtonCallback: (winid, cbfun) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return null;
		var prevcbfun = win.mouseButtonFunc;
		win.mouseButtonFunc = cbfun;
		return prevcbfun;
	},
	setCursorPosCallback: (winid, cbfun) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return null;
		var prevcbfun = win.cursorPosFunc;
		win.cursorPosFunc = cbfun;
		return prevcbfun;
	},
	setScrollCallback: (winid, cbfun) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return null;
		var prevcbfun = win.scrollFunc;
		win.scrollFunc = cbfun;
		return prevcbfun;
	},
	setDropCallback: (winid, cbfun) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return null;
		var prevcbfun = win.dropFunc;
		win.dropFunc = cbfun;
		return prevcbfun;
	},
	onDrop: (event) => {
		if (!GLFW.active || !GLFW.active.dropFunc) return;
		if (!event.dataTransfer || !event.dataTransfer.files || event.dataTransfer.files.length == 0)
			return;
		event.preventDefault();
		var filenames = _malloc(event.dataTransfer.files.length * 4);
		var filenamesArray = [];
		var count = event.dataTransfer.files.length;
		var written = 0;
		var drop_dir = '.glfw_dropped_files';
		FS.createPath('/', drop_dir);
		function save(file) {
			var path = '/' + drop_dir + '/' + file.name.replace(/\//g, '_');
			var reader = new FileReader();
			reader.onloadend = (e) => {
				if (reader.readyState != 2) {
					++written;
					out('failed to read dropped file: ' + file.name + ': ' + reader.error);
					return;
				}
				var data = e.target.result;
				FS.writeFile(path, new Uint8Array(data));
				if (++written === count) {
					((a1, a2, a3) => dynCall_viii.apply(null, [GLFW.active.dropFunc, a1, a2, a3]))(
						GLFW.active.id,
						count,
						filenames
					);
					for (var i = 0; i < filenamesArray.length; ++i) {
						_free(filenamesArray[i]);
					}
					_free(filenames);
				}
			};
			reader.readAsArrayBuffer(file);
			var filename = stringToNewUTF8(path);
			filenamesArray.push(filename);
			HEAPU32[(filenames + i * 4) >> 2] = filename;
		}
		for (var i = 0; i < count; ++i) {
			save(event.dataTransfer.files[i]);
		}
		return false;
	},
	onDragover: (event) => {
		if (!GLFW.active || !GLFW.active.dropFunc) return;
		event.preventDefault();
		return false;
	},
	setWindowSizeCallback: (winid, cbfun) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return null;
		var prevcbfun = win.windowSizeFunc;
		win.windowSizeFunc = cbfun;
		return prevcbfun;
	},
	setWindowCloseCallback: (winid, cbfun) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return null;
		var prevcbfun = win.windowCloseFunc;
		win.windowCloseFunc = cbfun;
		return prevcbfun;
	},
	setWindowRefreshCallback: (winid, cbfun) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return null;
		var prevcbfun = win.windowRefreshFunc;
		win.windowRefreshFunc = cbfun;
		return prevcbfun;
	},
	onClickRequestPointerLock: (e) => {
		if (!Browser.pointerLock && Module['canvas'].requestPointerLock) {
			Module['canvas'].requestPointerLock();
			e.preventDefault();
		}
	},
	setInputMode: (winid, mode, value) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return;
		switch (mode) {
			case 208897: {
				switch (value) {
					case 212993: {
						win.inputModes[mode] = value;
						Module['canvas'].removeEventListener('click', GLFW.onClickRequestPointerLock, true);
						Module['canvas'].exitPointerLock();
						break;
					}
					case 212994: {
						err('glfwSetInputMode called with GLFW_CURSOR_HIDDEN value not implemented');
						break;
					}
					case 212995: {
						win.inputModes[mode] = value;
						Module['canvas'].addEventListener('click', GLFW.onClickRequestPointerLock, true);
						Module['canvas'].requestPointerLock();
						break;
					}
					default: {
						err(`glfwSetInputMode called with unknown value parameter value: ${value}`);
						break;
					}
				}
				break;
			}
			case 208898: {
				err('glfwSetInputMode called with GLFW_STICKY_KEYS mode not implemented');
				break;
			}
			case 208899: {
				err('glfwSetInputMode called with GLFW_STICKY_MOUSE_BUTTONS mode not implemented');
				break;
			}
			case 208900: {
				err('glfwSetInputMode called with GLFW_LOCK_KEY_MODS mode not implemented');
				break;
			}
			case 3342341: {
				err('glfwSetInputMode called with GLFW_RAW_MOUSE_MOTION mode not implemented');
				break;
			}
			default: {
				err(`glfwSetInputMode called with unknown mode parameter value: ${mode}`);
				break;
			}
		}
	},
	getKey: (winid, key) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return 0;
		return win.keys[key];
	},
	getMouseButton: (winid, button) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return 0;
		return (win.buttons & (1 << button)) > 0;
	},
	getCursorPos: (winid, x, y) => {
		HEAPF64[x >> 3] = Browser.mouseX;
		HEAPF64[y >> 3] = Browser.mouseY;
	},
	getMousePos: (winid, x, y) => {
		HEAP32[x >> 2] = Browser.mouseX;
		HEAP32[y >> 2] = Browser.mouseY;
	},
	setCursorPos: (winid, x, y) => {},
	getWindowPos: (winid, x, y) => {
		var wx = 0;
		var wy = 0;
		var win = GLFW.WindowFromId(winid);
		if (win) {
			wx = win.x;
			wy = win.y;
		}
		if (x) {
			HEAP32[x >> 2] = wx;
		}
		if (y) {
			HEAP32[y >> 2] = wy;
		}
	},
	setWindowPos: (winid, x, y) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return;
		win.x = x;
		win.y = y;
	},
	getWindowSize: (winid, width, height) => {
		var ww = 0;
		var wh = 0;
		var win = GLFW.WindowFromId(winid);
		if (win) {
			ww = win.width;
			wh = win.height;
		}
		if (width) {
			HEAP32[width >> 2] = ww;
		}
		if (height) {
			HEAP32[height >> 2] = wh;
		}
	},
	setWindowSize: (winid, width, height) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return;
		if (GLFW.active.id == win.id) {
			Browser.setCanvasSize(width, height);
		}
	},
	defaultWindowHints: () => {
		GLFW.hints = Object.assign({}, GLFW.defaultHints);
	},
	createWindow: (width, height, title, monitor, share) => {
		var i, id;
		for (i = 0; i < GLFW.windows.length && GLFW.windows[i] !== null; i++) {}
		if (i > 0) throw 'glfwCreateWindow only supports one window at time currently';
		id = i + 1;
		if (width <= 0 || height <= 0) return 0;
		if (monitor) {
			Browser.requestFullscreen();
		} else {
			Browser.setCanvasSize(width, height);
		}
		for (i = 0; i < GLFW.windows.length && GLFW.windows[i] == null; i++) {}
		var useWebGL = GLFW.hints[139265] > 0;
		if (i == GLFW.windows.length) {
			if (useWebGL) {
				var contextAttributes = {
					antialias: GLFW.hints[135181] > 1,
					depth: GLFW.hints[135173] > 0,
					stencil: GLFW.hints[135174] > 0,
					alpha: GLFW.hints[135172] > 0
				};
				Module.ctx = Browser.createContext(Module['canvas'], true, true, contextAttributes);
			} else {
				Browser.init();
			}
		}
		if (!Module.ctx && useWebGL) return 0;
		const canvas = Module['canvas'];
		var win = new GLFW_Window(
			id,
			canvas.clientWidth,
			canvas.clientHeight,
			canvas.width,
			canvas.height,
			title,
			monitor,
			share
		);
		if (id - 1 == GLFW.windows.length) {
			GLFW.windows.push(win);
		} else {
			GLFW.windows[id - 1] = win;
		}
		GLFW.active = win;
		GLFW.adjustCanvasDimensions();
		return win.id;
	},
	destroyWindow: (winid) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return;
		if (win.windowCloseFunc) {
			((a1) => dynCall_vi.apply(null, [win.windowCloseFunc, a1]))(win.id);
		}
		GLFW.windows[win.id - 1] = null;
		if (GLFW.active.id == win.id) GLFW.active = null;
		for (var i = 0; i < GLFW.windows.length; i++) if (GLFW.windows[i] !== null) return;
		Module.ctx = Browser.destroyContext(Module['canvas'], true, true);
	},
	swapBuffers: (winid) => {},
	requestFullscreen(lockPointer, resizeCanvas) {
		Browser.lockPointer = lockPointer;
		Browser.resizeCanvas = resizeCanvas;
		if (typeof Browser.lockPointer == 'undefined') Browser.lockPointer = true;
		if (typeof Browser.resizeCanvas == 'undefined') Browser.resizeCanvas = false;
		var canvas = Module['canvas'];
		function fullscreenChange() {
			Browser.isFullscreen = false;
			var canvasContainer = canvas.parentNode;
			if (
				(document['fullscreenElement'] ||
					document['mozFullScreenElement'] ||
					document['msFullscreenElement'] ||
					document['webkitFullscreenElement'] ||
					document['webkitCurrentFullScreenElement']) === canvasContainer
			) {
				canvas.exitFullscreen = Browser.exitFullscreen;
				if (Browser.lockPointer) canvas.requestPointerLock();
				Browser.isFullscreen = true;
				if (Browser.resizeCanvas) {
					Browser.setFullscreenCanvasSize();
				} else {
					Browser.updateCanvasDimensions(canvas);
					Browser.updateResizeListeners();
				}
			} else {
				canvasContainer.parentNode.insertBefore(canvas, canvasContainer);
				canvasContainer.parentNode.removeChild(canvasContainer);
				if (Browser.resizeCanvas) {
					Browser.setWindowedCanvasSize();
				} else {
					Browser.updateCanvasDimensions(canvas);
					Browser.updateResizeListeners();
				}
			}
			if (Module['onFullScreen']) Module['onFullScreen'](Browser.isFullscreen);
			if (Module['onFullscreen']) Module['onFullscreen'](Browser.isFullscreen);
		}
		if (!Browser.fullscreenHandlersInstalled) {
			Browser.fullscreenHandlersInstalled = true;
			document.addEventListener('fullscreenchange', fullscreenChange, false);
			document.addEventListener('mozfullscreenchange', fullscreenChange, false);
			document.addEventListener('webkitfullscreenchange', fullscreenChange, false);
			document.addEventListener('MSFullscreenChange', fullscreenChange, false);
		}
		var canvasContainer = document.createElement('div');
		canvas.parentNode.insertBefore(canvasContainer, canvas);
		canvasContainer.appendChild(canvas);
		canvasContainer.requestFullscreen =
			canvasContainer['requestFullscreen'] ||
			canvasContainer['mozRequestFullScreen'] ||
			canvasContainer['msRequestFullscreen'] ||
			(canvasContainer['webkitRequestFullscreen']
				? () => canvasContainer['webkitRequestFullscreen'](Element['ALLOW_KEYBOARD_INPUT'])
				: null) ||
			(canvasContainer['webkitRequestFullScreen']
				? () => canvasContainer['webkitRequestFullScreen'](Element['ALLOW_KEYBOARD_INPUT'])
				: null);
		canvasContainer.requestFullscreen();
	},
	updateCanvasDimensions(canvas, wNative, hNative) {
		const scale = GLFW.getHiDPIScale();
		if (wNative && hNative) {
			canvas.widthNative = wNative;
			canvas.heightNative = hNative;
		} else {
			wNative = canvas.widthNative;
			hNative = canvas.heightNative;
		}
		var w = wNative;
		var h = hNative;
		if (Module['forcedAspectRatio'] && Module['forcedAspectRatio'] > 0) {
			if (w / h < Module['forcedAspectRatio']) {
				w = Math.round(h * Module['forcedAspectRatio']);
			} else {
				h = Math.round(w / Module['forcedAspectRatio']);
			}
		}
		if (
			(document['fullscreenElement'] ||
				document['mozFullScreenElement'] ||
				document['msFullscreenElement'] ||
				document['webkitFullscreenElement'] ||
				document['webkitCurrentFullScreenElement']) === canvas.parentNode &&
			typeof screen != 'undefined'
		) {
			var factor = Math.min(screen.width / w, screen.height / h);
			w = Math.round(w * factor);
			h = Math.round(h * factor);
		}
		if (Browser.resizeCanvas) {
			wNative = w;
			hNative = h;
		}
		const wNativeScaled = Math.floor(wNative * scale);
		const hNativeScaled = Math.floor(hNative * scale);
		if (canvas.width != wNativeScaled) canvas.width = wNativeScaled;
		if (canvas.height != hNativeScaled) canvas.height = hNativeScaled;
		if (typeof canvas.style != 'undefined') {
			if (wNativeScaled != wNative || hNativeScaled != hNative) {
				canvas.style.setProperty('width', wNative + 'px', 'important');
				canvas.style.setProperty('height', hNative + 'px', 'important');
			} else {
				canvas.style.removeProperty('width');
				canvas.style.removeProperty('height');
			}
		}
	},
	calculateMouseCoords(pageX, pageY) {
		var rect = Module['canvas'].getBoundingClientRect();
		var cw = Module['canvas'].clientWidth;
		var ch = Module['canvas'].clientHeight;
		var scrollX = typeof window.scrollX != 'undefined' ? window.scrollX : window.pageXOffset;
		var scrollY = typeof window.scrollY != 'undefined' ? window.scrollY : window.pageYOffset;
		assert(
			typeof scrollX != 'undefined' && typeof scrollY != 'undefined',
			'Unable to retrieve scroll position, mouse positions likely broken.'
		);
		var adjustedX = pageX - (scrollX + rect.left);
		var adjustedY = pageY - (scrollY + rect.top);
		adjustedX = adjustedX * (cw / rect.width);
		adjustedY = adjustedY * (ch / rect.height);
		return { x: adjustedX, y: adjustedY };
	},
	setWindowAttrib: (winid, attrib, value) => {
		var win = GLFW.WindowFromId(winid);
		if (!win) return;
		const isHiDPIAware = GLFW.isHiDPIAware();
		win.attributes[attrib] = value;
		if (isHiDPIAware !== GLFW.isHiDPIAware()) GLFW.adjustCanvasDimensions();
	},
	getDevicePixelRatio() {
		return (typeof devicePixelRatio == 'number' && devicePixelRatio) || 1;
	},
	isHiDPIAware() {
		if (GLFW.active) return GLFW.active.attributes[139276] > 0;
		else return false;
	},
	adjustCanvasDimensions() {
		const canvas = Module['canvas'];
		Browser.updateCanvasDimensions(canvas, canvas.clientWidth, canvas.clientHeight);
		Browser.updateResizeListeners();
	},
	getHiDPIScale() {
		return GLFW.isHiDPIAware() ? GLFW.scale : 1;
	},
	onDevicePixelRatioChange() {
		GLFW.onWindowContentScaleChanged(GLFW.getDevicePixelRatio());
		GLFW.adjustCanvasDimensions();
	},
	GLFW2ParamToGLFW3Param: (param) => {
		var table = {
			196609: 0,
			196610: 0,
			196611: 0,
			196612: 0,
			196613: 0,
			196614: 0,
			131073: 0,
			131074: 0,
			131075: 0,
			131076: 0,
			131077: 135169,
			131078: 135170,
			131079: 135171,
			131080: 135172,
			131081: 135173,
			131082: 135174,
			131083: 135183,
			131084: 135175,
			131085: 135176,
			131086: 135177,
			131087: 135178,
			131088: 135179,
			131089: 135180,
			131090: 0,
			131091: 135181,
			131092: 139266,
			131093: 139267,
			131094: 139270,
			131095: 139271,
			131096: 139272
		};
		return table[param];
	}
};
var _glfwCreateWindow = (width, height, title, monitor, share) =>
	GLFW.createWindow(width, height, title, monitor, share);
var _glfwDefaultWindowHints = () => GLFW.defaultWindowHints();
var _glfwDestroyWindow = (winid) => GLFW.destroyWindow(winid);
var _glfwGetPrimaryMonitor = () => 1;
var _glfwGetTime = () => GLFW.getTime() - GLFW.initialTime;
var _glfwGetVideoModes = (monitor, count) => {
	HEAP32[count >> 2] = 0;
	return 0;
};
var _glfwInit = () => {
	if (GLFW.windows) return 1;
	GLFW.initialTime = GLFW.getTime();
	GLFW.defaultWindowHints();
	GLFW.windows = new Array();
	GLFW.active = null;
	GLFW.scale = GLFW.getDevicePixelRatio();
	window.addEventListener('gamepadconnected', GLFW.onGamepadConnected, true);
	window.addEventListener('gamepaddisconnected', GLFW.onGamepadDisconnected, true);
	window.addEventListener('keydown', GLFW.onKeydown, true);
	window.addEventListener('keypress', GLFW.onKeyPress, true);
	window.addEventListener('keyup', GLFW.onKeyup, true);
	window.addEventListener('blur', GLFW.onBlur, true);
	GLFW.devicePixelRatioMQL = window.matchMedia(
		'(resolution: ' + GLFW.getDevicePixelRatio() + 'dppx)'
	);
	GLFW.devicePixelRatioMQL.addEventListener('change', GLFW.onDevicePixelRatioChange);
	Module['canvas'].addEventListener('touchmove', GLFW.onMousemove, true);
	Module['canvas'].addEventListener('touchstart', GLFW.onMouseButtonDown, true);
	Module['canvas'].addEventListener('touchcancel', GLFW.onMouseButtonUp, true);
	Module['canvas'].addEventListener('touchend', GLFW.onMouseButtonUp, true);
	Module['canvas'].addEventListener('mousemove', GLFW.onMousemove, true);
	Module['canvas'].addEventListener('mousedown', GLFW.onMouseButtonDown, true);
	Module['canvas'].addEventListener('mouseup', GLFW.onMouseButtonUp, true);
	Module['canvas'].addEventListener('wheel', GLFW.onMouseWheel, true);
	Module['canvas'].addEventListener('mousewheel', GLFW.onMouseWheel, true);
	Module['canvas'].addEventListener('mouseenter', GLFW.onMouseenter, true);
	Module['canvas'].addEventListener('mouseleave', GLFW.onMouseleave, true);
	Module['canvas'].addEventListener('drop', GLFW.onDrop, true);
	Module['canvas'].addEventListener('dragover', GLFW.onDragover, true);
	Browser.requestFullscreen = GLFW.requestFullscreen;
	Browser.calculateMouseCoords = GLFW.calculateMouseCoords;
	Browser.updateCanvasDimensions = GLFW.updateCanvasDimensions;
	Browser.resizeListeners.push((width, height) => {
		if (GLFW.isHiDPIAware()) {
			var canvas = Module['canvas'];
			GLFW.onCanvasResize(canvas.clientWidth, canvas.clientHeight, width, height);
		} else {
			GLFW.onCanvasResize(width, height, width, height);
		}
	});
	return 1;
};
var _glfwMakeContextCurrent = (winid) => {};
var _glfwSetCharCallback = (winid, cbfun) => GLFW.setCharCallback(winid, cbfun);
var _glfwSetCursorEnterCallback = (winid, cbfun) => {
	var win = GLFW.WindowFromId(winid);
	if (!win) return null;
	var prevcbfun = win.cursorEnterFunc;
	win.cursorEnterFunc = cbfun;
	return prevcbfun;
};
var _glfwSetCursorPosCallback = (winid, cbfun) => GLFW.setCursorPosCallback(winid, cbfun);
var _glfwSetDropCallback = (winid, cbfun) => GLFW.setDropCallback(winid, cbfun);
var _glfwSetErrorCallback = (cbfun) => {
	var prevcbfun = GLFW.errorFunc;
	GLFW.errorFunc = cbfun;
	return prevcbfun;
};
var _glfwSetKeyCallback = (winid, cbfun) => GLFW.setKeyCallback(winid, cbfun);
var _glfwSetMouseButtonCallback = (winid, cbfun) => GLFW.setMouseButtonCallback(winid, cbfun);
var _glfwSetScrollCallback = (winid, cbfun) => GLFW.setScrollCallback(winid, cbfun);
var _glfwSetWindowFocusCallback = (winid, cbfun) => {
	var win = GLFW.WindowFromId(winid);
	if (!win) return null;
	var prevcbfun = win.windowFocusFunc;
	win.windowFocusFunc = cbfun;
	return prevcbfun;
};
var _glfwSetWindowIconifyCallback = (winid, cbfun) => {
	var win = GLFW.WindowFromId(winid);
	if (!win) return null;
	var prevcbfun = win.windowIconifyFunc;
	win.windowIconifyFunc = cbfun;
	return prevcbfun;
};
var _glfwSetWindowShouldClose = (winid, value) => {
	var win = GLFW.WindowFromId(winid);
	if (!win) return;
	win.shouldClose = value;
};
var _glfwSetWindowSizeCallback = (winid, cbfun) => GLFW.setWindowSizeCallback(winid, cbfun);
var _glfwSwapBuffers = (winid) => GLFW.swapBuffers(winid);
var _glfwTerminate = () => {
	window.removeEventListener('gamepadconnected', GLFW.onGamepadConnected, true);
	window.removeEventListener('gamepaddisconnected', GLFW.onGamepadDisconnected, true);
	window.removeEventListener('keydown', GLFW.onKeydown, true);
	window.removeEventListener('keypress', GLFW.onKeyPress, true);
	window.removeEventListener('keyup', GLFW.onKeyup, true);
	window.removeEventListener('blur', GLFW.onBlur, true);
	Module['canvas'].removeEventListener('touchmove', GLFW.onMousemove, true);
	Module['canvas'].removeEventListener('touchstart', GLFW.onMouseButtonDown, true);
	Module['canvas'].removeEventListener('touchcancel', GLFW.onMouseButtonUp, true);
	Module['canvas'].removeEventListener('touchend', GLFW.onMouseButtonUp, true);
	Module['canvas'].removeEventListener('mousemove', GLFW.onMousemove, true);
	Module['canvas'].removeEventListener('mousedown', GLFW.onMouseButtonDown, true);
	Module['canvas'].removeEventListener('mouseup', GLFW.onMouseButtonUp, true);
	Module['canvas'].removeEventListener('wheel', GLFW.onMouseWheel, true);
	Module['canvas'].removeEventListener('mousewheel', GLFW.onMouseWheel, true);
	Module['canvas'].removeEventListener('mouseenter', GLFW.onMouseenter, true);
	Module['canvas'].removeEventListener('mouseleave', GLFW.onMouseleave, true);
	Module['canvas'].removeEventListener('drop', GLFW.onDrop, true);
	Module['canvas'].removeEventListener('dragover', GLFW.onDragover, true);
	if (GLFW.devicePixelRatioMQL)
		GLFW.devicePixelRatioMQL.removeEventListener('change', GLFW.onDevicePixelRatioChange);
	Module['canvas'].width = Module['canvas'].height = 1;
	GLFW.windows = null;
	GLFW.active = null;
};
var _glfwWindowHint = (target, hint) => {
	GLFW.hints[target] = hint;
};
var runAndAbortIfError = (func) => {
	try {
		return func();
	} catch (e) {
		abort(e);
	}
};
var Asyncify = {
	instrumentWasmImports(imports) {
		var importPattern = /^(invoke_.*|__asyncjs__.*)$/;
		for (let [x, original] of Object.entries(imports)) {
			let sig = original.sig;
			if (typeof original == 'function') {
				let isAsyncifyImport = original.isAsync || importPattern.test(x);
				imports[x] = function () {
					var originalAsyncifyState = Asyncify.state;
					try {
						return original.apply(null, arguments);
					} finally {
						var changedToDisabled =
							originalAsyncifyState === Asyncify.State.Normal &&
							Asyncify.state === Asyncify.State.Disabled;
						var ignoredInvoke = x.startsWith('invoke_') && true;
						if (
							Asyncify.state !== originalAsyncifyState &&
							!isAsyncifyImport &&
							!changedToDisabled &&
							!ignoredInvoke
						) {
							throw new Error(`import ${x} was not in ASYNCIFY_IMPORTS, but changed the state`);
						}
					}
				};
			}
		}
	},
	instrumentWasmExports(exports) {
		var ret = {};
		for (let [x, original] of Object.entries(exports)) {
			if (typeof original == 'function') {
				ret[x] = function () {
					Asyncify.exportCallStack.push(x);
					try {
						return original.apply(null, arguments);
					} finally {
						if (!ABORT) {
							var y = Asyncify.exportCallStack.pop();
							assert(y === x);
							Asyncify.maybeStopUnwind();
						}
					}
				};
			} else {
				ret[x] = original;
			}
		}
		return ret;
	},
	State: { Normal: 0, Unwinding: 1, Rewinding: 2, Disabled: 3 },
	state: 0,
	StackSize: 4096,
	currData: null,
	handleSleepReturnValue: 0,
	exportCallStack: [],
	callStackNameToId: {},
	callStackIdToName: {},
	callStackId: 0,
	asyncPromiseHandlers: null,
	sleepCallbacks: [],
	getCallStackId(funcName) {
		var id = Asyncify.callStackNameToId[funcName];
		if (id === undefined) {
			id = Asyncify.callStackId++;
			Asyncify.callStackNameToId[funcName] = id;
			Asyncify.callStackIdToName[id] = funcName;
		}
		return id;
	},
	maybeStopUnwind() {
		if (
			Asyncify.currData &&
			Asyncify.state === Asyncify.State.Unwinding &&
			Asyncify.exportCallStack.length === 0
		) {
			Asyncify.state = Asyncify.State.Normal;
			runtimeKeepalivePush();
			runAndAbortIfError(_asyncify_stop_unwind);
			if (typeof Fibers != 'undefined') {
				Fibers.trampoline();
			}
		}
	},
	whenDone() {
		assert(Asyncify.currData, 'Tried to wait for an async operation when none is in progress.');
		assert(
			!Asyncify.asyncPromiseHandlers,
			'Cannot have multiple async operations in flight at once'
		);
		return new Promise((resolve, reject) => {
			Asyncify.asyncPromiseHandlers = { resolve: resolve, reject: reject };
		});
	},
	allocateData() {
		var ptr = _malloc(12 + Asyncify.StackSize);
		Asyncify.setDataHeader(ptr, ptr + 12, Asyncify.StackSize);
		Asyncify.setDataRewindFunc(ptr);
		return ptr;
	},
	setDataHeader(ptr, stack, stackSize) {
		HEAPU32[ptr >> 2] = stack;
		HEAPU32[(ptr + 4) >> 2] = stack + stackSize;
	},
	setDataRewindFunc(ptr) {
		var bottomOfCallStack = Asyncify.exportCallStack[0];
		var rewindId = Asyncify.getCallStackId(bottomOfCallStack);
		HEAP32[(ptr + 8) >> 2] = rewindId;
	},
	getDataRewindFunc(ptr) {
		var id = HEAP32[(ptr + 8) >> 2];
		var name = Asyncify.callStackIdToName[id];
		var func = wasmExports[name];
		return func;
	},
	doRewind(ptr) {
		var start = Asyncify.getDataRewindFunc(ptr);
		runtimeKeepalivePop();
		return start();
	},
	handleSleep(startAsync) {
		assert(
			Asyncify.state !== Asyncify.State.Disabled,
			'Asyncify cannot be done during or after the runtime exits'
		);
		if (ABORT) return;
		if (Asyncify.state === Asyncify.State.Normal) {
			var reachedCallback = false;
			var reachedAfterCallback = false;
			startAsync((handleSleepReturnValue = 0) => {
				assert(
					!handleSleepReturnValue ||
						typeof handleSleepReturnValue == 'number' ||
						typeof handleSleepReturnValue == 'boolean'
				);
				if (ABORT) return;
				Asyncify.handleSleepReturnValue = handleSleepReturnValue;
				reachedCallback = true;
				if (!reachedAfterCallback) {
					return;
				}
				assert(
					!Asyncify.exportCallStack.length,
					'Waking up (starting to rewind) must be done from JS, without compiled code on the stack.'
				);
				Asyncify.state = Asyncify.State.Rewinding;
				runAndAbortIfError(() => _asyncify_start_rewind(Asyncify.currData));
				if (typeof Browser != 'undefined' && Browser.mainLoop.func) {
					Browser.mainLoop.resume();
				}
				var asyncWasmReturnValue,
					isError = false;
				try {
					asyncWasmReturnValue = Asyncify.doRewind(Asyncify.currData);
				} catch (err) {
					asyncWasmReturnValue = err;
					isError = true;
				}
				var handled = false;
				if (!Asyncify.currData) {
					var asyncPromiseHandlers = Asyncify.asyncPromiseHandlers;
					if (asyncPromiseHandlers) {
						Asyncify.asyncPromiseHandlers = null;
						(isError ? asyncPromiseHandlers.reject : asyncPromiseHandlers.resolve)(
							asyncWasmReturnValue
						);
						handled = true;
					}
				}
				if (isError && !handled) {
					throw asyncWasmReturnValue;
				}
			});
			reachedAfterCallback = true;
			if (!reachedCallback) {
				Asyncify.state = Asyncify.State.Unwinding;
				Asyncify.currData = Asyncify.allocateData();
				if (typeof Browser != 'undefined' && Browser.mainLoop.func) {
					Browser.mainLoop.pause();
				}
				runAndAbortIfError(() => _asyncify_start_unwind(Asyncify.currData));
			}
		} else if (Asyncify.state === Asyncify.State.Rewinding) {
			Asyncify.state = Asyncify.State.Normal;
			runAndAbortIfError(_asyncify_stop_rewind);
			_free(Asyncify.currData);
			Asyncify.currData = null;
			Asyncify.sleepCallbacks.forEach((func) => callUserCallback(func));
		} else {
			abort(`invalid state: ${Asyncify.state}`);
		}
		return Asyncify.handleSleepReturnValue;
	},
	handleAsync(startAsync) {
		return Asyncify.handleSleep((wakeUp) => {
			startAsync().then(wakeUp);
		});
	}
};
var getCFunc = (ident) => {
	var func = Module['_' + ident];
	assert(func, 'Cannot call unknown function ' + ident + ', make sure it is exported');
	return func;
};
var writeArrayToMemory = (array, buffer) => {
	assert(
		array.length >= 0,
		'writeArrayToMemory array must have a length (should be an array or typed array)'
	);
	HEAP8.set(array, buffer);
};
var stringToUTF8OnStack = (str) => {
	var size = lengthBytesUTF8(str) + 1;
	var ret = stackAlloc(size);
	stringToUTF8(str, ret, size);
	return ret;
};
var ccall = (ident, returnType, argTypes, args, opts) => {
	var toC = {
		string: (str) => {
			var ret = 0;
			if (str !== null && str !== undefined && str !== 0) {
				ret = stringToUTF8OnStack(str);
			}
			return ret;
		},
		array: (arr) => {
			var ret = stackAlloc(arr.length);
			writeArrayToMemory(arr, ret);
			return ret;
		}
	};
	function convertReturnValue(ret) {
		if (returnType === 'string') {
			return UTF8ToString(ret);
		}
		if (returnType === 'boolean') return Boolean(ret);
		return ret;
	}
	var func = getCFunc(ident);
	var cArgs = [];
	var stack = 0;
	assert(returnType !== 'array', 'Return type should not be "array".');
	if (args) {
		for (var i = 0; i < args.length; i++) {
			var converter = toC[argTypes[i]];
			if (converter) {
				if (stack === 0) stack = stackSave();
				cArgs[i] = converter(args[i]);
			} else {
				cArgs[i] = args[i];
			}
		}
	}
	var previousAsync = Asyncify.currData;
	var ret = func.apply(null, cArgs);
	function onDone(ret) {
		runtimeKeepalivePop();
		if (stack !== 0) stackRestore(stack);
		return convertReturnValue(ret);
	}
	var asyncMode = opts?.async;
	runtimeKeepalivePush();
	if (Asyncify.currData != previousAsync) {
		assert(
			!(previousAsync && Asyncify.currData),
			'We cannot start an async operation when one is already flight'
		);
		assert(!(previousAsync && !Asyncify.currData), 'We cannot stop an async operation in flight');
		assert(
			asyncMode,
			'The call to ' +
				ident +
				' is running asynchronously. If this was intended, add the async option to the ccall/cwrap call.'
		);
		return Asyncify.whenDone().then(onDone);
	}
	ret = onDone(ret);
	if (asyncMode) return Promise.resolve(ret);
	return ret;
};
var cwrap = (ident, returnType, argTypes, opts) =>
	function () {
		return ccall(ident, returnType, argTypes, arguments, opts);
	};
PThread.init();
var FSNode = function (parent, name, mode, rdev) {
	if (!parent) {
		parent = this;
	}
	this.parent = parent;
	this.mount = parent.mount;
	this.mounted = null;
	this.id = FS.nextInode++;
	this.name = name;
	this.mode = mode;
	this.node_ops = {};
	this.stream_ops = {};
	this.rdev = rdev;
};
var readMode = 292 | 73;
var writeMode = 146;
Object.defineProperties(FSNode.prototype, {
	read: {
		get: function () {
			return (this.mode & readMode) === readMode;
		},
		set: function (val) {
			val ? (this.mode |= readMode) : (this.mode &= ~readMode);
		}
	},
	write: {
		get: function () {
			return (this.mode & writeMode) === writeMode;
		},
		set: function (val) {
			val ? (this.mode |= writeMode) : (this.mode &= ~writeMode);
		}
	},
	isFolder: {
		get: function () {
			return FS.isDir(this.mode);
		}
	},
	isDevice: {
		get: function () {
			return FS.isChrdev(this.mode);
		}
	}
});
FS.FSNode = FSNode;
FS.createPreloadedFile = FS_createPreloadedFile;
FS.staticInit();
Module['FS_createPath'] = FS.createPath;
Module['FS_createDataFile'] = FS.createDataFile;
Module['FS_createPreloadedFile'] = FS.createPreloadedFile;
Module['FS_unlink'] = FS.unlink;
Module['FS_createLazyFile'] = FS.createLazyFile;
Module['FS_createDevice'] = FS.createDevice;
var GLctx;
for (var i = 0; i < 32; ++i) tempFixedLengthArray.push(new Array(i));
var miniTempWebGLFloatBuffersStorage = new Float32Array(288);
for (var i = 0; i < 288; ++i) {
	miniTempWebGLFloatBuffers[i] = miniTempWebGLFloatBuffersStorage.subarray(0, i + 1);
}
var miniTempWebGLIntBuffersStorage = new Int32Array(288);
for (var i = 0; i < 288; ++i) {
	miniTempWebGLIntBuffers[i] = miniTempWebGLIntBuffersStorage.subarray(0, i + 1);
}
Module['requestFullscreen'] = Browser.requestFullscreen;
Module['requestFullScreen'] = Browser.requestFullScreen;
Module['requestAnimationFrame'] = Browser.requestAnimationFrame;
Module['setCanvasSize'] = Browser.setCanvasSize;
Module['pauseMainLoop'] = Browser.mainLoop.pause;
Module['resumeMainLoop'] = Browser.mainLoop.resume;
Module['getUserMedia'] = Browser.getUserMedia;
Module['createContext'] = Browser.createContext;
var preloadedImages = {};
var preloadedAudios = {};
var proxiedFunctionTable = [
	_proc_exit,
	exitOnMainThread,
	pthreadCreateProxied,
	___syscall_faccessat,
	___syscall_fcntl64,
	___syscall_getcwd,
	___syscall_ioctl,
	___syscall_openat,
	_emscripten_get_element_css_size,
	_emscripten_get_gamepad_status,
	_emscripten_get_num_gamepads,
	_emscripten_sample_gamepad_data,
	setCanvasElementSizeMainThread,
	_emscripten_set_click_callback_on_thread,
	_emscripten_set_fullscreenchange_callback_on_thread,
	_emscripten_set_gamepadconnected_callback_on_thread,
	_emscripten_set_gamepaddisconnected_callback_on_thread,
	_emscripten_set_pointerlockchange_callback_on_thread,
	_emscripten_set_resize_callback_on_thread,
	_emscripten_set_touchcancel_callback_on_thread,
	_emscripten_set_touchend_callback_on_thread,
	_emscripten_set_touchmove_callback_on_thread,
	_emscripten_set_touchstart_callback_on_thread,
	_emscripten_set_window_title,
	_fd_close,
	_fd_read,
	_fd_seek,
	_fd_write
];
function checkIncomingModuleAPI() {
	ignoredModuleProp('fetchSettings');
}
var wasmImports = {
	GetWindowInnerHeight: GetWindowInnerHeight,
	GetWindowInnerWidth: GetWindowInnerWidth,
	__assert_fail: ___assert_fail,
	__cxa_throw: ___cxa_throw,
	__emscripten_init_main_thread_js: ___emscripten_init_main_thread_js,
	__emscripten_thread_cleanup: ___emscripten_thread_cleanup,
	__pthread_create_js: ___pthread_create_js,
	__syscall_faccessat: ___syscall_faccessat,
	__syscall_fcntl64: ___syscall_fcntl64,
	__syscall_getcwd: ___syscall_getcwd,
	__syscall_ioctl: ___syscall_ioctl,
	__syscall_openat: ___syscall_openat,
	_emscripten_get_now_is_monotonic: __emscripten_get_now_is_monotonic,
	_emscripten_notify_mailbox_postmessage: __emscripten_notify_mailbox_postmessage,
	_emscripten_receive_on_main_thread_js: __emscripten_receive_on_main_thread_js,
	_emscripten_thread_mailbox_await: __emscripten_thread_mailbox_await,
	_emscripten_thread_set_strongref: __emscripten_thread_set_strongref,
	abort: _abort,
	emscripten_asm_const_int: _emscripten_asm_const_int,
	emscripten_check_blocking_allowed: _emscripten_check_blocking_allowed,
	emscripten_date_now: _emscripten_date_now,
	emscripten_exit_with_live_runtime: _emscripten_exit_with_live_runtime,
	emscripten_get_element_css_size: _emscripten_get_element_css_size,
	emscripten_get_gamepad_status: _emscripten_get_gamepad_status,
	emscripten_get_heap_max: _emscripten_get_heap_max,
	emscripten_get_now: _emscripten_get_now,
	emscripten_get_num_gamepads: _emscripten_get_num_gamepads,
	emscripten_glActiveTexture: _emscripten_glActiveTexture,
	emscripten_glAttachShader: _emscripten_glAttachShader,
	emscripten_glBeginQueryEXT: _emscripten_glBeginQueryEXT,
	emscripten_glBindAttribLocation: _emscripten_glBindAttribLocation,
	emscripten_glBindBuffer: _emscripten_glBindBuffer,
	emscripten_glBindFramebuffer: _emscripten_glBindFramebuffer,
	emscripten_glBindRenderbuffer: _emscripten_glBindRenderbuffer,
	emscripten_glBindTexture: _emscripten_glBindTexture,
	emscripten_glBindVertexArrayOES: _emscripten_glBindVertexArrayOES,
	emscripten_glBlendColor: _emscripten_glBlendColor,
	emscripten_glBlendEquation: _emscripten_glBlendEquation,
	emscripten_glBlendEquationSeparate: _emscripten_glBlendEquationSeparate,
	emscripten_glBlendFunc: _emscripten_glBlendFunc,
	emscripten_glBlendFuncSeparate: _emscripten_glBlendFuncSeparate,
	emscripten_glBufferData: _emscripten_glBufferData,
	emscripten_glBufferSubData: _emscripten_glBufferSubData,
	emscripten_glCheckFramebufferStatus: _emscripten_glCheckFramebufferStatus,
	emscripten_glClear: _emscripten_glClear,
	emscripten_glClearColor: _emscripten_glClearColor,
	emscripten_glClearDepthf: _emscripten_glClearDepthf,
	emscripten_glClearStencil: _emscripten_glClearStencil,
	emscripten_glColorMask: _emscripten_glColorMask,
	emscripten_glCompileShader: _emscripten_glCompileShader,
	emscripten_glCompressedTexImage2D: _emscripten_glCompressedTexImage2D,
	emscripten_glCompressedTexSubImage2D: _emscripten_glCompressedTexSubImage2D,
	emscripten_glCopyTexImage2D: _emscripten_glCopyTexImage2D,
	emscripten_glCopyTexSubImage2D: _emscripten_glCopyTexSubImage2D,
	emscripten_glCreateProgram: _emscripten_glCreateProgram,
	emscripten_glCreateShader: _emscripten_glCreateShader,
	emscripten_glCullFace: _emscripten_glCullFace,
	emscripten_glDeleteBuffers: _emscripten_glDeleteBuffers,
	emscripten_glDeleteFramebuffers: _emscripten_glDeleteFramebuffers,
	emscripten_glDeleteProgram: _emscripten_glDeleteProgram,
	emscripten_glDeleteQueriesEXT: _emscripten_glDeleteQueriesEXT,
	emscripten_glDeleteRenderbuffers: _emscripten_glDeleteRenderbuffers,
	emscripten_glDeleteShader: _emscripten_glDeleteShader,
	emscripten_glDeleteTextures: _emscripten_glDeleteTextures,
	emscripten_glDeleteVertexArraysOES: _emscripten_glDeleteVertexArraysOES,
	emscripten_glDepthFunc: _emscripten_glDepthFunc,
	emscripten_glDepthMask: _emscripten_glDepthMask,
	emscripten_glDepthRangef: _emscripten_glDepthRangef,
	emscripten_glDetachShader: _emscripten_glDetachShader,
	emscripten_glDisable: _emscripten_glDisable,
	emscripten_glDisableVertexAttribArray: _emscripten_glDisableVertexAttribArray,
	emscripten_glDrawArrays: _emscripten_glDrawArrays,
	emscripten_glDrawArraysInstancedANGLE: _emscripten_glDrawArraysInstancedANGLE,
	emscripten_glDrawBuffersWEBGL: _emscripten_glDrawBuffersWEBGL,
	emscripten_glDrawElements: _emscripten_glDrawElements,
	emscripten_glDrawElementsInstancedANGLE: _emscripten_glDrawElementsInstancedANGLE,
	emscripten_glEnable: _emscripten_glEnable,
	emscripten_glEnableVertexAttribArray: _emscripten_glEnableVertexAttribArray,
	emscripten_glEndQueryEXT: _emscripten_glEndQueryEXT,
	emscripten_glFinish: _emscripten_glFinish,
	emscripten_glFlush: _emscripten_glFlush,
	emscripten_glFramebufferRenderbuffer: _emscripten_glFramebufferRenderbuffer,
	emscripten_glFramebufferTexture2D: _emscripten_glFramebufferTexture2D,
	emscripten_glFrontFace: _emscripten_glFrontFace,
	emscripten_glGenBuffers: _emscripten_glGenBuffers,
	emscripten_glGenFramebuffers: _emscripten_glGenFramebuffers,
	emscripten_glGenQueriesEXT: _emscripten_glGenQueriesEXT,
	emscripten_glGenRenderbuffers: _emscripten_glGenRenderbuffers,
	emscripten_glGenTextures: _emscripten_glGenTextures,
	emscripten_glGenVertexArraysOES: _emscripten_glGenVertexArraysOES,
	emscripten_glGenerateMipmap: _emscripten_glGenerateMipmap,
	emscripten_glGetActiveAttrib: _emscripten_glGetActiveAttrib,
	emscripten_glGetActiveUniform: _emscripten_glGetActiveUniform,
	emscripten_glGetAttachedShaders: _emscripten_glGetAttachedShaders,
	emscripten_glGetAttribLocation: _emscripten_glGetAttribLocation,
	emscripten_glGetBooleanv: _emscripten_glGetBooleanv,
	emscripten_glGetBufferParameteriv: _emscripten_glGetBufferParameteriv,
	emscripten_glGetError: _emscripten_glGetError,
	emscripten_glGetFloatv: _emscripten_glGetFloatv,
	emscripten_glGetFramebufferAttachmentParameteriv:
		_emscripten_glGetFramebufferAttachmentParameteriv,
	emscripten_glGetIntegerv: _emscripten_glGetIntegerv,
	emscripten_glGetProgramInfoLog: _emscripten_glGetProgramInfoLog,
	emscripten_glGetProgramiv: _emscripten_glGetProgramiv,
	emscripten_glGetQueryObjecti64vEXT: _emscripten_glGetQueryObjecti64vEXT,
	emscripten_glGetQueryObjectivEXT: _emscripten_glGetQueryObjectivEXT,
	emscripten_glGetQueryObjectui64vEXT: _emscripten_glGetQueryObjectui64vEXT,
	emscripten_glGetQueryObjectuivEXT: _emscripten_glGetQueryObjectuivEXT,
	emscripten_glGetQueryivEXT: _emscripten_glGetQueryivEXT,
	emscripten_glGetRenderbufferParameteriv: _emscripten_glGetRenderbufferParameteriv,
	emscripten_glGetShaderInfoLog: _emscripten_glGetShaderInfoLog,
	emscripten_glGetShaderPrecisionFormat: _emscripten_glGetShaderPrecisionFormat,
	emscripten_glGetShaderSource: _emscripten_glGetShaderSource,
	emscripten_glGetShaderiv: _emscripten_glGetShaderiv,
	emscripten_glGetString: _emscripten_glGetString,
	emscripten_glGetTexParameterfv: _emscripten_glGetTexParameterfv,
	emscripten_glGetTexParameteriv: _emscripten_glGetTexParameteriv,
	emscripten_glGetUniformLocation: _emscripten_glGetUniformLocation,
	emscripten_glGetUniformfv: _emscripten_glGetUniformfv,
	emscripten_glGetUniformiv: _emscripten_glGetUniformiv,
	emscripten_glGetVertexAttribPointerv: _emscripten_glGetVertexAttribPointerv,
	emscripten_glGetVertexAttribfv: _emscripten_glGetVertexAttribfv,
	emscripten_glGetVertexAttribiv: _emscripten_glGetVertexAttribiv,
	emscripten_glHint: _emscripten_glHint,
	emscripten_glIsBuffer: _emscripten_glIsBuffer,
	emscripten_glIsEnabled: _emscripten_glIsEnabled,
	emscripten_glIsFramebuffer: _emscripten_glIsFramebuffer,
	emscripten_glIsProgram: _emscripten_glIsProgram,
	emscripten_glIsQueryEXT: _emscripten_glIsQueryEXT,
	emscripten_glIsRenderbuffer: _emscripten_glIsRenderbuffer,
	emscripten_glIsShader: _emscripten_glIsShader,
	emscripten_glIsTexture: _emscripten_glIsTexture,
	emscripten_glIsVertexArrayOES: _emscripten_glIsVertexArrayOES,
	emscripten_glLineWidth: _emscripten_glLineWidth,
	emscripten_glLinkProgram: _emscripten_glLinkProgram,
	emscripten_glPixelStorei: _emscripten_glPixelStorei,
	emscripten_glPolygonOffset: _emscripten_glPolygonOffset,
	emscripten_glQueryCounterEXT: _emscripten_glQueryCounterEXT,
	emscripten_glReadPixels: _emscripten_glReadPixels,
	emscripten_glReleaseShaderCompiler: _emscripten_glReleaseShaderCompiler,
	emscripten_glRenderbufferStorage: _emscripten_glRenderbufferStorage,
	emscripten_glSampleCoverage: _emscripten_glSampleCoverage,
	emscripten_glScissor: _emscripten_glScissor,
	emscripten_glShaderBinary: _emscripten_glShaderBinary,
	emscripten_glShaderSource: _emscripten_glShaderSource,
	emscripten_glStencilFunc: _emscripten_glStencilFunc,
	emscripten_glStencilFuncSeparate: _emscripten_glStencilFuncSeparate,
	emscripten_glStencilMask: _emscripten_glStencilMask,
	emscripten_glStencilMaskSeparate: _emscripten_glStencilMaskSeparate,
	emscripten_glStencilOp: _emscripten_glStencilOp,
	emscripten_glStencilOpSeparate: _emscripten_glStencilOpSeparate,
	emscripten_glTexImage2D: _emscripten_glTexImage2D,
	emscripten_glTexParameterf: _emscripten_glTexParameterf,
	emscripten_glTexParameterfv: _emscripten_glTexParameterfv,
	emscripten_glTexParameteri: _emscripten_glTexParameteri,
	emscripten_glTexParameteriv: _emscripten_glTexParameteriv,
	emscripten_glTexSubImage2D: _emscripten_glTexSubImage2D,
	emscripten_glUniform1f: _emscripten_glUniform1f,
	emscripten_glUniform1fv: _emscripten_glUniform1fv,
	emscripten_glUniform1i: _emscripten_glUniform1i,
	emscripten_glUniform1iv: _emscripten_glUniform1iv,
	emscripten_glUniform2f: _emscripten_glUniform2f,
	emscripten_glUniform2fv: _emscripten_glUniform2fv,
	emscripten_glUniform2i: _emscripten_glUniform2i,
	emscripten_glUniform2iv: _emscripten_glUniform2iv,
	emscripten_glUniform3f: _emscripten_glUniform3f,
	emscripten_glUniform3fv: _emscripten_glUniform3fv,
	emscripten_glUniform3i: _emscripten_glUniform3i,
	emscripten_glUniform3iv: _emscripten_glUniform3iv,
	emscripten_glUniform4f: _emscripten_glUniform4f,
	emscripten_glUniform4fv: _emscripten_glUniform4fv,
	emscripten_glUniform4i: _emscripten_glUniform4i,
	emscripten_glUniform4iv: _emscripten_glUniform4iv,
	emscripten_glUniformMatrix2fv: _emscripten_glUniformMatrix2fv,
	emscripten_glUniformMatrix3fv: _emscripten_glUniformMatrix3fv,
	emscripten_glUniformMatrix4fv: _emscripten_glUniformMatrix4fv,
	emscripten_glUseProgram: _emscripten_glUseProgram,
	emscripten_glValidateProgram: _emscripten_glValidateProgram,
	emscripten_glVertexAttrib1f: _emscripten_glVertexAttrib1f,
	emscripten_glVertexAttrib1fv: _emscripten_glVertexAttrib1fv,
	emscripten_glVertexAttrib2f: _emscripten_glVertexAttrib2f,
	emscripten_glVertexAttrib2fv: _emscripten_glVertexAttrib2fv,
	emscripten_glVertexAttrib3f: _emscripten_glVertexAttrib3f,
	emscripten_glVertexAttrib3fv: _emscripten_glVertexAttrib3fv,
	emscripten_glVertexAttrib4f: _emscripten_glVertexAttrib4f,
	emscripten_glVertexAttrib4fv: _emscripten_glVertexAttrib4fv,
	emscripten_glVertexAttribDivisorANGLE: _emscripten_glVertexAttribDivisorANGLE,
	emscripten_glVertexAttribPointer: _emscripten_glVertexAttribPointer,
	emscripten_glViewport: _emscripten_glViewport,
	emscripten_num_logical_cores: _emscripten_num_logical_cores,
	emscripten_resize_heap: _emscripten_resize_heap,
	emscripten_sample_gamepad_data: _emscripten_sample_gamepad_data,
	emscripten_set_canvas_element_size: _emscripten_set_canvas_element_size,
	emscripten_set_click_callback_on_thread: _emscripten_set_click_callback_on_thread,
	emscripten_set_fullscreenchange_callback_on_thread:
		_emscripten_set_fullscreenchange_callback_on_thread,
	emscripten_set_gamepadconnected_callback_on_thread:
		_emscripten_set_gamepadconnected_callback_on_thread,
	emscripten_set_gamepaddisconnected_callback_on_thread:
		_emscripten_set_gamepaddisconnected_callback_on_thread,
	emscripten_set_main_loop: _emscripten_set_main_loop,
	emscripten_set_pointerlockchange_callback_on_thread:
		_emscripten_set_pointerlockchange_callback_on_thread,
	emscripten_set_resize_callback_on_thread: _emscripten_set_resize_callback_on_thread,
	emscripten_set_touchcancel_callback_on_thread: _emscripten_set_touchcancel_callback_on_thread,
	emscripten_set_touchend_callback_on_thread: _emscripten_set_touchend_callback_on_thread,
	emscripten_set_touchmove_callback_on_thread: _emscripten_set_touchmove_callback_on_thread,
	emscripten_set_touchstart_callback_on_thread: _emscripten_set_touchstart_callback_on_thread,
	emscripten_set_window_title: _emscripten_set_window_title,
	exit: _exit,
	fd_close: _fd_close,
	fd_read: _fd_read,
	fd_seek: _fd_seek,
	fd_write: _fd_write,
	getentropy: _getentropy,
	glActiveTexture: _glActiveTexture,
	glAttachShader: _glAttachShader,
	glBindAttribLocation: _glBindAttribLocation,
	glBindBuffer: _glBindBuffer,
	glBindTexture: _glBindTexture,
	glBlendFunc: _glBlendFunc,
	glBufferData: _glBufferData,
	glBufferSubData: _glBufferSubData,
	glClear: _glClear,
	glClearColor: _glClearColor,
	glClearDepthf: _glClearDepthf,
	glCompileShader: _glCompileShader,
	glCompressedTexImage2D: _glCompressedTexImage2D,
	glCreateProgram: _glCreateProgram,
	glCreateShader: _glCreateShader,
	glCullFace: _glCullFace,
	glDeleteBuffers: _glDeleteBuffers,
	glDeleteProgram: _glDeleteProgram,
	glDeleteShader: _glDeleteShader,
	glDeleteTextures: _glDeleteTextures,
	glDepthFunc: _glDepthFunc,
	glDetachShader: _glDetachShader,
	glDisable: _glDisable,
	glDisableVertexAttribArray: _glDisableVertexAttribArray,
	glDrawArrays: _glDrawArrays,
	glDrawElements: _glDrawElements,
	glEnable: _glEnable,
	glEnableVertexAttribArray: _glEnableVertexAttribArray,
	glFrontFace: _glFrontFace,
	glGenBuffers: _glGenBuffers,
	glGenTextures: _glGenTextures,
	glGetAttribLocation: _glGetAttribLocation,
	glGetFloatv: _glGetFloatv,
	glGetProgramInfoLog: _glGetProgramInfoLog,
	glGetProgramiv: _glGetProgramiv,
	glGetShaderInfoLog: _glGetShaderInfoLog,
	glGetShaderiv: _glGetShaderiv,
	glGetString: _glGetString,
	glGetUniformLocation: _glGetUniformLocation,
	glLinkProgram: _glLinkProgram,
	glPixelStorei: _glPixelStorei,
	glReadPixels: _glReadPixels,
	glShaderSource: _glShaderSource,
	glTexImage2D: _glTexImage2D,
	glTexParameterf: _glTexParameterf,
	glTexParameteri: _glTexParameteri,
	glUniform1i: _glUniform1i,
	glUniform4f: _glUniform4f,
	glUniformMatrix4fv: _glUniformMatrix4fv,
	glUseProgram: _glUseProgram,
	glVertexAttribPointer: _glVertexAttribPointer,
	glViewport: _glViewport,
	glfwCreateWindow: _glfwCreateWindow,
	glfwDefaultWindowHints: _glfwDefaultWindowHints,
	glfwDestroyWindow: _glfwDestroyWindow,
	glfwGetPrimaryMonitor: _glfwGetPrimaryMonitor,
	glfwGetTime: _glfwGetTime,
	glfwGetVideoModes: _glfwGetVideoModes,
	glfwInit: _glfwInit,
	glfwMakeContextCurrent: _glfwMakeContextCurrent,
	glfwSetCharCallback: _glfwSetCharCallback,
	glfwSetCursorEnterCallback: _glfwSetCursorEnterCallback,
	glfwSetCursorPosCallback: _glfwSetCursorPosCallback,
	glfwSetDropCallback: _glfwSetDropCallback,
	glfwSetErrorCallback: _glfwSetErrorCallback,
	glfwSetKeyCallback: _glfwSetKeyCallback,
	glfwSetMouseButtonCallback: _glfwSetMouseButtonCallback,
	glfwSetScrollCallback: _glfwSetScrollCallback,
	glfwSetWindowFocusCallback: _glfwSetWindowFocusCallback,
	glfwSetWindowIconifyCallback: _glfwSetWindowIconifyCallback,
	glfwSetWindowShouldClose: _glfwSetWindowShouldClose,
	glfwSetWindowSizeCallback: _glfwSetWindowSizeCallback,
	glfwSwapBuffers: _glfwSwapBuffers,
	glfwTerminate: _glfwTerminate,
	glfwWindowHint: _glfwWindowHint,
	memory: wasmMemory
};
Asyncify.instrumentWasmImports(wasmImports);
var wasmExports = createWasm();
var ___wasm_call_ctors = createExportWrapper('__wasm_call_ctors');
var _main = (Module['_main'] = createExportWrapper('main'));
var _free = (Module['_free'] = createExportWrapper('free'));
var _malloc = (Module['_malloc'] = createExportWrapper('malloc'));
var _fflush = (Module['_fflush'] = createExportWrapper('fflush'));
var ___errno_location = createExportWrapper('__errno_location');
var __emscripten_tls_init = (Module['__emscripten_tls_init'] =
	createExportWrapper('_emscripten_tls_init'));
var _pthread_self = (Module['_pthread_self'] = () =>
	(_pthread_self = Module['_pthread_self'] = wasmExports['pthread_self'])());
var __emscripten_run_callback_on_thread = createExportWrapper('_emscripten_run_callback_on_thread');
var __emscripten_thread_init = (Module['__emscripten_thread_init'] =
	createExportWrapper('_emscripten_thread_init'));
var __emscripten_thread_crashed = (Module['__emscripten_thread_crashed'] = createExportWrapper(
	'_emscripten_thread_crashed'
));
var _emscripten_main_thread_process_queued_calls = createExportWrapper(
	'emscripten_main_thread_process_queued_calls'
);
var _emscripten_main_runtime_thread_id = createExportWrapper('emscripten_main_runtime_thread_id');
var _emscripten_stack_get_base = () =>
	(_emscripten_stack_get_base = wasmExports['emscripten_stack_get_base'])();
var _emscripten_stack_get_end = () =>
	(_emscripten_stack_get_end = wasmExports['emscripten_stack_get_end'])();
var __emscripten_run_on_main_thread_js = createExportWrapper('_emscripten_run_on_main_thread_js');
var __emscripten_thread_free_data = createExportWrapper('_emscripten_thread_free_data');
var __emscripten_thread_exit = (Module['__emscripten_thread_exit'] =
	createExportWrapper('_emscripten_thread_exit'));
var __emscripten_check_mailbox = createExportWrapper('_emscripten_check_mailbox');
var _emscripten_stack_init = () =>
	(_emscripten_stack_init = wasmExports['emscripten_stack_init'])();
var _emscripten_stack_set_limits = (a0, a1) =>
	(_emscripten_stack_set_limits = wasmExports['emscripten_stack_set_limits'])(a0, a1);
var _emscripten_stack_get_free = () =>
	(_emscripten_stack_get_free = wasmExports['emscripten_stack_get_free'])();
var stackSave = createExportWrapper('stackSave');
var stackRestore = createExportWrapper('stackRestore');
var stackAlloc = createExportWrapper('stackAlloc');
var _emscripten_stack_get_current = () =>
	(_emscripten_stack_get_current = wasmExports['emscripten_stack_get_current'])();
var ___cxa_is_pointer_type = createExportWrapper('__cxa_is_pointer_type');
var dynCall_v = (Module['dynCall_v'] = createExportWrapper('dynCall_v'));
var dynCall_vi = (Module['dynCall_vi'] = createExportWrapper('dynCall_vi'));
var dynCall_viii = (Module['dynCall_viii'] = createExportWrapper('dynCall_viii'));
var dynCall_ii = (Module['dynCall_ii'] = createExportWrapper('dynCall_ii'));
var dynCall_iii = (Module['dynCall_iii'] = createExportWrapper('dynCall_iii'));
var dynCall_vii = (Module['dynCall_vii'] = createExportWrapper('dynCall_vii'));
var dynCall_viiiii = (Module['dynCall_viiiii'] = createExportWrapper('dynCall_viiiii'));
var dynCall_viiii = (Module['dynCall_viiii'] = createExportWrapper('dynCall_viiii'));
var dynCall_vidd = (Module['dynCall_vidd'] = createExportWrapper('dynCall_vidd'));
var dynCall_iiii = (Module['dynCall_iiii'] = createExportWrapper('dynCall_iiii'));
var dynCall_vffff = (Module['dynCall_vffff'] = createExportWrapper('dynCall_vffff'));
var dynCall_vf = (Module['dynCall_vf'] = createExportWrapper('dynCall_vf'));
var dynCall_viiiiiiii = (Module['dynCall_viiiiiiii'] = createExportWrapper('dynCall_viiiiiiii'));
var dynCall_viiiiiiiii = (Module['dynCall_viiiiiiiii'] = createExportWrapper('dynCall_viiiiiiiii'));
var dynCall_i = (Module['dynCall_i'] = createExportWrapper('dynCall_i'));
var dynCall_vff = (Module['dynCall_vff'] = createExportWrapper('dynCall_vff'));
var dynCall_viiiiiii = (Module['dynCall_viiiiiii'] = createExportWrapper('dynCall_viiiiiii'));
var dynCall_vfi = (Module['dynCall_vfi'] = createExportWrapper('dynCall_vfi'));
var dynCall_viif = (Module['dynCall_viif'] = createExportWrapper('dynCall_viif'));
var dynCall_vif = (Module['dynCall_vif'] = createExportWrapper('dynCall_vif'));
var dynCall_viff = (Module['dynCall_viff'] = createExportWrapper('dynCall_viff'));
var dynCall_vifff = (Module['dynCall_vifff'] = createExportWrapper('dynCall_vifff'));
var dynCall_viffff = (Module['dynCall_viffff'] = createExportWrapper('dynCall_viffff'));
var dynCall_viiiiii = (Module['dynCall_viiiiii'] = createExportWrapper('dynCall_viiiiii'));
var dynCall_jiji = (Module['dynCall_jiji'] = createExportWrapper('dynCall_jiji'));
var dynCall_iidiiii = (Module['dynCall_iidiiii'] = createExportWrapper('dynCall_iidiiii'));
var _asyncify_start_unwind = createExportWrapper('asyncify_start_unwind');
var _asyncify_stop_unwind = createExportWrapper('asyncify_stop_unwind');
var _asyncify_start_rewind = createExportWrapper('asyncify_start_rewind');
var _asyncify_stop_rewind = createExportWrapper('asyncify_stop_rewind');
var ___start_em_js = (Module['___start_em_js'] = 39983);
var ___stop_em_js = (Module['___stop_em_js'] = 40056);
Module['addRunDependency'] = addRunDependency;
Module['removeRunDependency'] = removeRunDependency;
Module['FS_createPath'] = FS.createPath;
Module['FS_createLazyFile'] = FS.createLazyFile;
Module['FS_createDevice'] = FS.createDevice;
Module['wasmMemory'] = wasmMemory;
Module['keepRuntimeAlive'] = keepRuntimeAlive;
Module['ccall'] = ccall;
Module['cwrap'] = cwrap;
Module['ExitStatus'] = ExitStatus;
Module['FS_createPreloadedFile'] = FS.createPreloadedFile;
Module['FS_createDataFile'] = FS.createDataFile;
Module['FS_unlink'] = FS.unlink;
var missingLibrarySymbols = [
	'writeI53ToI64Clamped',
	'writeI53ToI64Signaling',
	'writeI53ToU64Clamped',
	'writeI53ToU64Signaling',
	'convertI32PairToI53',
	'convertU32PairToI53',
	'growMemory',
	'isLeapYear',
	'ydayFromDate',
	'arraySum',
	'addDays',
	'inetPton4',
	'inetNtop4',
	'inetPton6',
	'inetNtop6',
	'readSockaddr',
	'writeSockaddr',
	'getHostByName',
	'getCallstack',
	'emscriptenLog',
	'convertPCtoSourceLocation',
	'runMainThreadEmAsm',
	'jstoi_s',
	'getExecutableName',
	'listenOnce',
	'autoResumeAudioContext',
	'dynCallLegacy',
	'getDynCaller',
	'dynCall',
	'setWasmTableEntry',
	'getWasmTableEntry',
	'asmjsMangle',
	'handleAllocatorInit',
	'HandleAllocator',
	'getNativeTypeSize',
	'STACK_SIZE',
	'STACK_ALIGN',
	'POINTER_SIZE',
	'ASSERTIONS',
	'uleb128Encode',
	'generateFuncType',
	'convertJsFunctionToWasm',
	'getEmptyTableSlot',
	'updateTableMap',
	'getFunctionAddress',
	'addFunction',
	'removeFunction',
	'reallyNegative',
	'unSign',
	'strLen',
	'reSign',
	'formatString',
	'intArrayToString',
	'AsciiToString',
	'stringToAscii',
	'UTF16ToString',
	'stringToUTF16',
	'lengthBytesUTF16',
	'UTF32ToString',
	'stringToUTF32',
	'lengthBytesUTF32',
	'registerKeyEventCallback',
	'registerWheelEventCallback',
	'registerFocusEventCallback',
	'fillDeviceOrientationEventData',
	'registerDeviceOrientationEventCallback',
	'fillDeviceMotionEventData',
	'registerDeviceMotionEventCallback',
	'screenOrientation',
	'fillOrientationChangeEventData',
	'registerOrientationChangeEventCallback',
	'JSEvents_requestFullscreen',
	'JSEvents_resizeCanvasForFullscreen',
	'registerRestoreOldStyle',
	'hideEverythingExceptGivenElement',
	'restoreHiddenElements',
	'setLetterbox',
	'softFullscreenResizeWebGLRenderTarget',
	'doRequestFullscreen',
	'registerPointerlockErrorEventCallback',
	'requestPointerLock',
	'fillVisibilityChangeEventData',
	'registerVisibilityChangeEventCallback',
	'registerBeforeUnloadEventCallback',
	'fillBatteryEventData',
	'battery',
	'registerBatteryEventCallback',
	'setCanvasElementSize',
	'getCanvasSizeCallingThread',
	'getCanvasSizeMainThread',
	'getCanvasElementSize',
	'jsStackTrace',
	'stackTrace',
	'getEnvStrings',
	'checkWasiClock',
	'wasiRightsToMuslOFlags',
	'wasiOFlagsToMuslOFlags',
	'createDyncallWrapper',
	'setImmediateWrapped',
	'clearImmediateWrapped',
	'polyfillSetImmediate',
	'getPromise',
	'makePromise',
	'idsToPromises',
	'makePromiseCallback',
	'findMatchingCatch',
	'Browser_asyncPrepareDataCounter',
	'getSocketFromFD',
	'getSocketAddress',
	'FS_mkdirTree',
	'_setNetworkCallback',
	'writeGLArray',
	'emscripten_webgl_destroy_context_before_on_calling_thread',
	'registerWebGlEventCallback',
	'SDL_unicode',
	'SDL_ttfContext',
	'SDL_audio',
	'ALLOC_NORMAL',
	'ALLOC_STACK',
	'allocate',
	'writeStringToMemory',
	'writeAsciiToMemory'
];
missingLibrarySymbols.forEach(missingLibrarySymbol);
var unexportedSymbols = [
	'run',
	'addOnPreRun',
	'addOnInit',
	'addOnPreMain',
	'addOnExit',
	'addOnPostRun',
	'FS_createFolder',
	'FS_createLink',
	'FS_readFile',
	'out',
	'err',
	'callMain',
	'abort',
	'wasmExports',
	'stackAlloc',
	'stackSave',
	'stackRestore',
	'getTempRet0',
	'setTempRet0',
	'writeStackCookie',
	'checkStackCookie',
	'writeI53ToI64',
	'readI53FromI64',
	'readI53FromU64',
	'convertI32PairToI53Checked',
	'ptrToString',
	'zeroMemory',
	'exitJS',
	'getHeapMax',
	'abortOnCannotGrowMemory',
	'ENV',
	'MONTH_DAYS_REGULAR',
	'MONTH_DAYS_LEAP',
	'MONTH_DAYS_REGULAR_CUMULATIVE',
	'MONTH_DAYS_LEAP_CUMULATIVE',
	'ERRNO_CODES',
	'ERRNO_MESSAGES',
	'setErrNo',
	'DNS',
	'Protocols',
	'Sockets',
	'initRandomFill',
	'randomFill',
	'timers',
	'warnOnce',
	'UNWIND_CACHE',
	'readEmAsmArgsArray',
	'readEmAsmArgs',
	'runEmAsmFunction',
	'jstoi_q',
	'handleException',
	'runtimeKeepalivePush',
	'runtimeKeepalivePop',
	'callUserCallback',
	'maybeExit',
	'asyncLoad',
	'alignMemory',
	'mmapAlloc',
	'wasmTable',
	'noExitRuntime',
	'getCFunc',
	'sigToWasmTypes',
	'freeTableIndexes',
	'functionsInTableMap',
	'setValue',
	'getValue',
	'PATH',
	'PATH_FS',
	'UTF8Decoder',
	'UTF8ArrayToString',
	'UTF8ToString',
	'stringToUTF8Array',
	'stringToUTF8',
	'lengthBytesUTF8',
	'intArrayFromString',
	'UTF16Decoder',
	'stringToNewUTF8',
	'stringToUTF8OnStack',
	'writeArrayToMemory',
	'JSEvents',
	'specialHTMLTargets',
	'maybeCStringToJsString',
	'findEventTarget',
	'findCanvasEventTarget',
	'getBoundingClientRect',
	'fillMouseEventData',
	'registerMouseEventCallback',
	'registerUiEventCallback',
	'fillFullscreenChangeEventData',
	'registerFullscreenChangeEventCallback',
	'currentFullscreenStrategy',
	'restoreOldWindowedStyle',
	'fillPointerlockChangeEventData',
	'registerPointerlockChangeEventCallback',
	'registerTouchEventCallback',
	'fillGamepadEventData',
	'registerGamepadEventCallback',
	'disableGamepadApiIfItThrows',
	'setCanvasElementSizeCallingThread',
	'setCanvasElementSizeMainThread',
	'demangle',
	'demangleAll',
	'doReadv',
	'doWritev',
	'safeSetTimeout',
	'promiseMap',
	'uncaughtExceptionCount',
	'exceptionLast',
	'exceptionCaught',
	'ExceptionInfo',
	'Browser',
	'setMainLoop',
	'wget',
	'SYSCALLS',
	'preloadPlugins',
	'FS_modeStringToFlags',
	'FS_getMode',
	'FS_stdin_getChar_buffer',
	'FS_stdin_getChar',
	'FS',
	'MEMFS',
	'TTY',
	'PIPEFS',
	'SOCKFS',
	'tempFixedLengthArray',
	'miniTempWebGLFloatBuffers',
	'miniTempWebGLIntBuffers',
	'heapObjectForWebGLType',
	'heapAccessShiftForWebGLHeap',
	'webgl_enable_ANGLE_instanced_arrays',
	'webgl_enable_OES_vertex_array_object',
	'webgl_enable_WEBGL_draw_buffers',
	'webgl_enable_WEBGL_multi_draw',
	'GL',
	'emscriptenWebGLGet',
	'computeUnpackAlignedImageSize',
	'colorChannelsInGlTextureFormat',
	'emscriptenWebGLGetTexPixelData',
	'__glGenObject',
	'emscriptenWebGLGetUniform',
	'webglGetUniformLocation',
	'webglPrepareUniformLocationsBeforeFirstUse',
	'webglGetLeftBracePos',
	'emscriptenWebGLGetVertexAttrib',
	'__glGetActiveAttribOrUniform',
	'emscripten_webgl_power_preferences',
	'AL',
	'GLUT',
	'EGL',
	'GLEW',
	'IDBStore',
	'runAndAbortIfError',
	'Asyncify',
	'Fibers',
	'SDL',
	'SDL_gfx',
	'GLFW_Window',
	'GLFW',
	'allocateUTF8',
	'allocateUTF8OnStack',
	'PThread',
	'terminateWorker',
	'killThread',
	'cleanupThread',
	'registerTLSInit',
	'cancelThread',
	'spawnThread',
	'exitOnMainThread',
	'proxyToMainThread',
	'proxiedJSCallArgs',
	'invokeEntryPoint',
	'checkMailbox'
];
unexportedSymbols.forEach(unexportedRuntimeSymbol);
var calledRun;
dependenciesFulfilled = function runCaller() {
	if (!calledRun) run();
	if (!calledRun) dependenciesFulfilled = runCaller;
};
function callMain() {
	assert(
		runDependencies == 0,
		'cannot call main when async dependencies remain! (listen on Module["onRuntimeInitialized"])'
	);
	assert(__ATPRERUN__.length == 0, 'cannot call main when preRun functions remain to be called');
	var entryFunction = _main;
	var argc = 0;
	var argv = 0;
	try {
		var ret = entryFunction(argc, argv);
		exitJS(ret, true);
		return ret;
	} catch (e) {
		return handleException(e);
	}
}
function stackCheckInit() {
	assert(!ENVIRONMENT_IS_PTHREAD);
	_emscripten_stack_init();
	writeStackCookie();
}
function run() {
	if (runDependencies > 0) {
		return;
	}
	if (!ENVIRONMENT_IS_PTHREAD) stackCheckInit();
	if (ENVIRONMENT_IS_PTHREAD) {
		initRuntime();
		startWorker(Module);
		return;
	}
	preRun();
	if (runDependencies > 0) {
		return;
	}
	function doRun() {
		if (calledRun) return;
		calledRun = true;
		Module['calledRun'] = true;
		if (ABORT) return;
		initRuntime();
		preMain();
		if (Module['onRuntimeInitialized']) Module['onRuntimeInitialized']();
		if (shouldRunNow) callMain();
		postRun();
	}
	if (Module['setStatus']) {
		Module['setStatus']('Running...');
		setTimeout(function () {
			setTimeout(function () {
				Module['setStatus']('');
			}, 1);
			doRun();
		}, 1);
	} else {
		doRun();
	}
	checkStackCookie();
}
function checkUnflushedContent() {
	var oldOut = out;
	var oldErr = err;
	var has = false;
	out = err = (x) => {
		has = true;
	};
	try {
		_fflush(0);
		['stdout', 'stderr'].forEach(function (name) {
			var info = FS.analyzePath('/dev/' + name);
			if (!info) return;
			var stream = info.object;
			var rdev = stream.rdev;
			var tty = TTY.ttys[rdev];
			if (tty?.output?.length) {
				has = true;
			}
		});
	} catch (e) {}
	out = oldOut;
	err = oldErr;
	if (has) {
		warnOnce(
			'stdio streams had content in them that was not flushed. you should set EXIT_RUNTIME to 1 (see the Emscripten FAQ), or make sure to emit a newline when you printf etc.'
		);
	}
}
if (Module['preInit']) {
	if (typeof Module['preInit'] == 'function') Module['preInit'] = [Module['preInit']];
	while (Module['preInit'].length > 0) {
		Module['preInit'].pop()();
	}
}
var shouldRunNow = true;
if (Module['noInitialRun']) shouldRunNow = false;
run();
