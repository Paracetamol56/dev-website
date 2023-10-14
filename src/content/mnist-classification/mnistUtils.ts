
export const getModelArchitecture = (name: string) => {
  let model = `model = tf.sequential( ) \n`

  if (name === 'ANN') {
      model +=`model.add( tf.layers.flatten( { inputShape: [28, 28, 1] } ) ) \n`
      model +=`model.add( tf.layers.dense( { units: 42, activation: 'relu' } ) ) \n`
      model +=`model.add( tf.layers.dense( { units: 10, activation: 'softmax' } ) ) \n`    
  } else if (name === 'CNN') {
      model +=`model.add( tf.layers.conv2d( { kernelSize: 3, filters: 16, activation: 'relu' , inputShape: [28, 28, 1]} ) ) \n`
      model +=`model.add( tf.layers.conv2d( { kernelSize: 3, filters: 16, activation: 'relu' } ) ) \n`
      model +=`model.add( tf.layers.maxPooling2d( { poolSize: 2, strides: 2 } ) ) \n`
      model +=`model.add( tf.layers.flatten( { } ) ) \n`
      model +=`model.add( tf.layers.dense( { units: 10, activation: 'softmax' } ) ) \n`
  }

  model +=`\n`
  model +=`const myOptim = 'rmsprop' \n`        
  model +=`model.compile( { loss: 'categoricalCrossentropy', optimizer: myOptim, metrics:['accuracy'] } ) \n`

  return model
}