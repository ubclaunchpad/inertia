import { combineReducers, createStore, applyMiddleware } from 'redux';
import thunk from 'redux-thunk';
import Dashboard from './dashboard';
import Main from './main';
import Auth from './auth';

const rootReducer = combineReducers({
  Dashboard,
  Main,
  Auth,
});

const store = createStore(
  rootReducer,
  applyMiddleware(thunk)
);

export default store;
