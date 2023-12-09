import { MongoClient } from 'mongodb';
import { env } from '$env/dynamic/private';

const uri = env.DB_URI || 'mongodb://localhost:27017/dev';
const client = new MongoClient(uri);
client.connect();
export default client.db('dev');
