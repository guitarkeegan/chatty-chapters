export default function LoginForm() {
  return (
    <form className="flex flex-col gap-4 items-center border-black border pt-20 pb-12 px-4 font-body">
      <label htmlFor="email">Email</label>
      <input
        className="bg-main border border-black"
        name="email"
        type="email"
      />
      <label htmlFor="password">Password</label>
      <input
        className="bg-main border border-black"
        name="password"
        type="password"
      />
      <button className="px-4 mt-4 border border-black" type="submit">
        Enter
      </button>
    </form>
  );
}
