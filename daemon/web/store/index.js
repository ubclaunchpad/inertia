import { combineReducers, createStore, applyMiddleware } from 'redux';
import thunk from 'redux-thunk';
import Dashboard from './dashboard';
import Main from './main';
import Login from './login';


const rootReducer = combineReducers({
  Dashboard,
  Main,
  Login,
});

const store = createStore(
  rootReducer,
  applyMiddleware(thunk)
);

export default store;
