import LoginForm from "@/components/login-form";

export default function Home() {
  return (
    <main>
      <section className="bg-main h-screen flex flex-col gap-24 justify-center items-center">
        <div>
          <h2 className="font-display text-4xl text-center">Music Tutor</h2>
        </div>
        <LoginForm />
      </section>
      {/* <h1 className="font-display text-2xl">Music Tutor</h1>
      <p className="font-body">I am the body</p> */}
    </main>
  );
}
