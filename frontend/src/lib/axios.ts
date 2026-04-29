import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

// On any 401, fire a custom event so AuthContext can react without a circular import
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (axios.isAxiosError(err) && err.response?.status === 401) {
      window.dispatchEvent(new CustomEvent('auth:unauthorized'))
    }
    return Promise.reject(err)
  },
)

export default api
