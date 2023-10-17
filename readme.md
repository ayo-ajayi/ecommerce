# Ecommerce API

## Routes
- **all users**
  - [x] POST /api/singup
  - [x] POST /api/verify
  - [x] POST /api/login
  - [x] POST /api/forgot-password
  - [x] POST /api/reset-password
  - [x] POST  /api/refresh-token
  - [x] POST /api/resend-verification-otp
  - [x] GET api/category/:id
  - [x] GET api/categories
  - [x] GET api/item/:id
  - [x] GET api/items
  - [x] GET api/review/:id
  - [x] GET api/search?q=

  - **authenticated users**
    - [x] POST api/logout
    - [x] GET api/profile
    - [x] POST api/add-address
    - [x] DELETE api/remove-address/:id
    - [x] GET api/addresses
    - [x] GET api/address/:id
    - [x] PUT api/update-address/:id
  
    - **admin**
      - [x] GET api/admin/users
      - [x] POST api/admin/create-category
      - [x] PUT api/admin/update-category/:id
      - [x] DELETE api/admin/delete-category/:id
    - **vendor**
      - [x] DELETE api/vendor/delete-item/:id
      - [x] PUT api/vendor/update-item/:id
      - [x] POST api/vendor/create-item
    - **customer**
      - [x] PUT api/customer/update-cart
      - [x] GET api/customer/cart
      - [x] POST api/customer/post-review 


### User:

- **SignUp**:
  - User receives a mail with OTP.
  - User is redirected to a page where they can enter the OTP.
  - Upon success, the user is redirected to the login page.

- **Login**:
  - Since the user just signed up, the email is prefilled.
  - User enters the password and logs in.
  - Access token and refresh tokens are generated upon successful login.

- **Logout**:
  - Access token is deleted from the database.

- **Refresh Token**:
  - Refresh token is used to generate new access token and refresh token.
  - User needs to login again if the refresh token is expired.

- **Forgot Password**:
  - User enters email and receives OTP by mail.
  - User is redirected to a page where they can enter the OTP.
  - Upon success, the user is redirected to a page where they can enter a new password.

- **Reset Password**:
  - User enters a new password and confirms it.

### User Types:

There are 3 types of Users:

1. **Admin**
2. **Vendor**
3. **Customer**

- User submits a form to become a vendor.
- They submit their details and upload their documents for verification by the admin.
- Once the admin verifies the documents, the user becomes a vendor and receives a notification by mail.
- Upon signing in, the vendor is asked if they want to log in as a vendor or customer.
- If they choose vendor, they are redirected to the vendor dashboard. If they choose customer, they are redirected to the customer dashboard.
- The vendor can switch between the two dashboards.
- A vendor can be blocked by the admin. If blocked, the vendor can only log in as a customer.

### Communication:

- Users can chat with the admin/support.
- Admin can chat with vendors about issues, orders, or items.

### Item:

- Only vendors can add an item.
- Vendors supply the item details and upload the item images.
- Vendors can decide to supply various categories for their item or not. If they don't, the item is added to the default category.

### Category:

- Only admin can add a category.
- Admin supplies the category details and uploads the category image.

### Review:

- A vendor cannot submit a review for their own item.
- A customer can submit review as a star rating and a comment or either of them.
- A customer can only update their review. They cannot have multiple reviews for the same item.
- A customer can submit a review for an item only if they have ordered the item.
- Users can decide to submit reviews anonymously or not.

### Cards:

- Users can add and remove cards from their account.
- Small card images can be displayed for the cards with the details hidden.

### Order:

- Orders can be placed anonymously by a customer who is not logged in.
- Upon submitting an order, the customer is redirected to the payment gateway.
- Once payment is successful, the order is placed, and both the admin and vendor are notified by mail and on the dashboard.
- Orders are tracked by the customer, vendor, and admin and updated accordingly.

### Cart:

- Users can add and remove items from their cart.
- Anonymous users need to have a persistent cart, which can be stored in their browser's local storage.







### License

[MIT](LICENSE)


###  Author

-   [Ayomide Ajayi](https://github.com/ayo-ajayi)