import {
  TEST_MAIN_ACTION,
} from './_constants';

export const mainAction = payload => (dispatch) => {
  dispatch({
    type: TEST_MAIN_ACTION,
    payload,
  });
};
