
export const getModelArchitecture = (name: string) => {
  let modelCode = `let newModel = tf.sequential( ) \n`

  if (name === 'ANN') {
    modelCode +=`newModel.add( tf.layers.flatten( { inputShape: [28, 28, 1] } ) ) \n`
    modelCode +=`newModel.add( tf.layers.dense( { units: 42, activation: 'relu' } ) ) \n`
    modelCode +=`newModel.add( tf.layers.dense( { units: 10, activation: 'softmax' } ) ) \n`    
  } else if (name === 'CNN') {
    modelCode +=`newModel.add( tf.layers.conv2d( { kernelSize: 3, filters: 16, activation: 'relu' , inputShape: [28, 28, 1]} ) ) \n`
    modelCode +=`newModel.add( tf.layers.conv2d( { kernelSize: 3, filters: 16, activation: 'relu' } ) ) \n`
    modelCode +=`newModel.add( tf.layers.maxPooling2d( { poolSize: 2, strides: 2 } ) ) \n`
    modelCode +=`newModel.add( tf.layers.flatten( { } ) ) \n`
    modelCode +=`newModel.add( tf.layers.dense( { units: 10, activation: 'softmax' } ) ) \n`
  }

  modelCode +=`\n`
  modelCode +=`const myOptim = 'rmsprop' \n`        
  modelCode +=`newModel.compile( { loss: 'categoricalCrossentropy', optimizer: myOptim, metrics:['accuracy'] } ) \n`

  return modelCode
};

