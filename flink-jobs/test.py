from pyflink.datastream import StreamExecutionEnvironment

env = StreamExecutionEnvironment.get_execution_environment()

env.from_collection(["hello", "world"]).print()

env.execute("test")