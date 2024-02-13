'use strict';
var Module = {};
var ENVIRONMENT_IS_NODE =
	typeof process == 'object' &&
	typeof process.versions == 'object' &&
	typeof process.versions.node == 'string';
if (ENVIRONMENT_IS_NODE) {
	var nodeWorkerThreads = require('worker_threads');
	var parentPort = nodeWorkerThreads.parentPort;
	parentPort.on('message', (data) => onmessage({ data: data }));
	var fs = require('fs');
	var vm = require('vm');
	Object.assign(global, {
		self: global,
		require: require,
		Module: Module,
		location: { href: __filename },
		Worker: nodeWorkerThreads.Worker,
		importScripts: (f) => vm.runInThisContext(fs.readFileSync(f, 'utf8'), { filename: f }),
		postMessage: (msg) => parentPort.postMessage(msg),
		performance: global.performance || { now: Date.now }
	});
}
var initializedJS = false;
function assert(condition, text) {
	if (!condition) abort('Assertion failed: ' + text);
}
function threadPrintErr() {
	var text = Array.prototype.slice.call(arguments).join(' ');
	if (ENVIRONMENT_IS_NODE) {
		fs.writeSync(2, text + '\n');
		return;
	}
	console.error(text);
}
function threadAlert() {
	var text = Array.prototype.slice.call(arguments).join(' ');
	postMessage({ cmd: 'alert', text: text, threadId: Module['_pthread_self']() });
}
var out = () => {
	throw 'out() is not defined in worker.js.';
};
var err = threadPrintErr;
self.alert = threadAlert;
var dbg = threadPrintErr;
Module['instantiateWasm'] = (info, receiveInstance) => {
	var module = Module['wasmModule'];
	Module['wasmModule'] = null;
	var instance = new WebAssembly.Instance(module, info);
	return receiveInstance(instance);
};
self.onunhandledrejection = (e) => {
	throw e.reason || e;
};
function handleMessage(e) {
	try {
		if (e.data.cmd === 'load') {
			let messageQueue = [];
			self.onmessage = (e) => messageQueue.push(e);
			self.startWorker = (instance) => {
				postMessage({ cmd: 'loaded' });
				for (let msg of messageQueue) {
					handleMessage(msg);
				}
				self.onmessage = handleMessage;
			};
			Module['wasmModule'] = e.data.wasmModule;
			for (const handler of e.data.handlers) {
				Module[handler] = (...args) => {
					postMessage({ cmd: 'callHandler', handler: handler, args: args });
				};
			}
			Module['wasmMemory'] = e.data.wasmMemory;
			Module['buffer'] = Module['wasmMemory'].buffer;
			Module['workerID'] = e.data.workerID;
			Module['ENVIRONMENT_IS_PTHREAD'] = true;
			if (typeof e.data.urlOrBlob == 'string') {
				importScripts(e.data.urlOrBlob);
			} else {
				var objectUrl = URL.createObjectURL(e.data.urlOrBlob);
				importScripts(objectUrl);
				URL.revokeObjectURL(objectUrl);
			}
		} else if (e.data.cmd === 'run') {
			Module['__emscripten_thread_init'](e.data.pthread_ptr, 0, 0, 1);
			Module['__emscripten_thread_mailbox_await'](e.data.pthread_ptr);
			assert(e.data.pthread_ptr);
			Module['establishStackSpace']();
			Module['PThread'].receiveObjectTransfer(e.data);
			Module['PThread'].threadInitTLS();
			if (!initializedJS) {
				initializedJS = true;
			}
			try {
				Module['invokeEntryPoint'](e.data.start_routine, e.data.arg);
			} catch (ex) {
				if (ex != 'unwind') {
					throw ex;
				}
			}
		} else if (e.data.cmd === 'cancel') {
			if (Module['_pthread_self']()) {
				Module['__emscripten_thread_exit'](-1);
			}
		} else if (e.data.target === 'setimmediate') {
		} else if (e.data.cmd === 'checkMailbox') {
			if (initializedJS) {
				Module['checkMailbox']();
			}
		} else if (e.data.cmd) {
			err(`worker.js received unknown command ${e.data.cmd}`);
			err(e.data);
		}
	} catch (ex) {
		err(`worker.js onmessage() captured an uncaught exception: ${ex}`);
		if (ex?.stack) err(ex.stack);
		Module['__emscripten_thread_crashed']?.();
		throw ex;
	}
}
self.onmessage = handleMessage;
